package pipe

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/server/config"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/message"
	"github.com/laincloud/entry/server/models"
	"github.com/laincloud/entry/server/term"
	"github.com/laincloud/entry/server/util"
)

const (
	aliveDecectionInterval = 10 * time.Second
	feedbackTimeout        = 100 * time.Millisecond
)

// Pipe is a full duplex channel between the docker container and the terminal
type Pipe struct {
	conn           *websocket.Conn
	marshal        util.Marshaler
	requestBuffer  chan []byte
	responseBuffer chan []byte
	session        *models.Session
	unMarshal      util.Unmarshaler
	wg             *sync.WaitGroup
	writeLock      *sync.Mutex
}

// NewPipe return an initialized *Pipe
func NewPipe(conn *websocket.Conn, marshal util.Marshaler, session *models.Session, unMarshal util.Unmarshaler, wg *sync.WaitGroup, writeLock *sync.Mutex) *Pipe {
	return &Pipe{
		conn:           conn,
		marshal:        marshal,
		requestBuffer:  make(chan []byte, 1),
		responseBuffer: make(chan []byte, 1),
		session:        session,
		unMarshal:      unMarshal,
		wg:             wg,
		writeLock:      writeLock,
	}
}

// HandleRequest handle request from the client
func (p *Pipe) HandleRequest(execID string, sessionWriter io.WriteCloser, g *global.Global) {
	var (
		err   error
		wsMsg []byte
		buf   bytes.Buffer
	)
	time.Sleep(time.Second)
	inMsg := message.RequestMessage{}
	for err == nil {
		if _, wsMsg, err = p.conn.ReadMessage(); err == nil {
			if unmarshalErr := p.unMarshal(wsMsg, &inMsg); unmarshalErr == nil {
				switch inMsg.MsgType {
				case message.RequestMessage_PLAIN:
					if _, err = sessionWriter.Write(inMsg.Content); err != nil {
						log.Errorf("sessionWriter.Write() failed, inMsg.Content: %s(%v), error: %s, session: %+v.", inMsg.Content, inMsg.Content, err, p.session)
						continue
					}

					err = p.handleInput(inMsg.Content, &buf, g)
				case message.RequestMessage_WINCH:
					if width, height := util.GetWidthAndHeight(inMsg.Content); width >= 0 && height >= 0 {
						err = g.DockerClient.ResizeExecTTY(execID, height, width)
					}
				}

			} else {
				log.Errorf("Unmarshall request failed, error: %s, session: %+v.", unmarshalErr.Error(), p.session)
			}
		}
	}
	if err != nil {
		log.Errorf("handleRequest failed, error: %s, session: %+v.", err.Error(), p.session)
	}

	sessionWriter.Close()
	p.wg.Done()
}

// HandleResponse handle response from the container
func (p *Pipe) HandleResponse(respType message.ResponseMessage_ResponseType, sessionReader io.ReadCloser, sessionReplay *SessionReplay) {
	var (
		err  error
		size int
	)
	buf := make([]byte, config.WriteBufferSize)
	cursor := 0
	for err == nil {
		if size, err = sessionReader.Read(buf[cursor:]); err == nil || (err == io.EOF && size > 0) {
			validLen := util.GetValidUT8Length(buf[:cursor+size])
			if validLen == 0 {
				log.Errorf("No valid UTF8 sequence prefix")
				break
			}

			p.feedbackInput(buf[:validLen])

			if sessionReplay != nil {
				sessionReplay.record(buf[:validLen])
			}

			outMsg := &message.ResponseMessage{
				MsgType: respType,
				Content: buf[:validLen],
			}
			data, marshalErr := p.marshal(outMsg)
			if marshalErr == nil {
				p.writeLock.Lock()
				err = p.conn.WriteMessage(websocket.BinaryMessage, data)
				p.writeLock.Unlock()
				cursor := size - validLen
				for i := 0; i < cursor; i++ {
					buf[i] = buf[cursor+i]
				}
			} else {
				log.Errorf("Marshal response error: %s", marshalErr.Error())
			}
		}
	}
	if err != nil {
		log.Errorf("HandleResponse ended: %s", err.Error())
	}

	sessionReader.Close()
	p.wg.Done()
}

// HandleAliveDetection handle alive detection for the websocket
func (p *Pipe) HandleAliveDetection(isStop <-chan int) {
	pingMsg := &message.ResponseMessage{
		MsgType: message.ResponseMessage_PING,
		Content: []byte("ping"),
	}
	data, _ := p.marshal(pingMsg)
	ticker := time.NewTicker(aliveDecectionInterval)
	defer ticker.Stop()
	for {
		select {
		case <-isStop:
			return
		case <-ticker.C:
			p.writeLock.Lock()
			p.conn.WriteMessage(websocket.BinaryMessage, data)
			p.writeLock.Unlock()
		}
	}
}

func (p *Pipe) askForFeedback(input []byte) []byte {
	p.requestBuffer <- input
	select {
	case feedback := <-p.responseBuffer:
		return feedback
	case <-time.After(feedbackTimeout):
		log.Error("askForFeedback() timeout, will give up.")
		select {
		case <-p.requestBuffer:
		default:
		}
		return []byte{}
	}
}

// feedback respond to tab completion, up arrow or down arrow
func (p *Pipe) feedbackInput(feedback []byte) {
	select {
	case input := <-p.requestBuffer:
		var escapedFeedback []byte
		switch {
		case term.IsTab(input):
			escapedFeedback = term.EscapeTabCompletion(feedback)
		case term.HasUpArrowSuffix(input) || term.HasDownArrowSuffix(input):
			escapedFeedback = term.EscapeHistoryCommand(feedback)
		default:
			log.Errorf("Unknown input: %v, will return empty feedback.", input)
		}
		p.responseBuffer <- escapedFeedback
	default:
	}
}

func (p *Pipe) handleInput(input []byte, buf *bytes.Buffer, g *global.Global) error {
	switch {
	case term.IsCR(input):
		p.saveCommand(buf.Bytes(), g)
		buf.Reset()
	case term.IsTab(input):
		buf.Write(p.askForFeedback(input))
	default:
		if _, err := buf.Write(input); err != nil {
			log.Errorf("buf.Write() failed, error: %s, session: %+v.", err, p.session)
			return err
		}

		log.Infof("buf.Write() succeed, inMsg.Content: %v, session: %+v.", input, p.session)
		switch {
		case term.HasUpArrowSuffix(buf.Bytes()):
			feedback := p.askForFeedback(buf.Bytes())
			buf.Truncate(len(buf.Bytes()) - len(term.UpArrow))
			buf.Write(feedback)
		case term.HasDownArrowSuffix(buf.Bytes()):
			feedback := p.askForFeedback(buf.Bytes())
			buf.Truncate(len(buf.Bytes()) - len(term.DownArrow))
			buf.Write(feedback)
		}
	}

	return nil
}

func (p *Pipe) saveCommand(input []byte, g *global.Global) {
	commandContent := string(term.EscapeInput(input))
	if commandContent != "" {
		command := models.Command{
			SessionID: p.session.SessionID,
			User:      p.session.User,
			Content:   commandContent,
		}
		g.DB.Create(&command)
		if command.IsRisky() {
			log.Warnf("Dangerous command! Will alert entry owners... Command.Content: %v, session: %+v.", command.Content, p.session)
			go func() {
				if err1 := command.Alert(*p.session, g); err1 != nil {
					log.Errorf("command.Alert() failed, error: %v.", err1)
				}
			}()
		} else {
			log.Infof("command.Content: %v, session: %+v.", command.Content, p.session)
		}
	}
}
