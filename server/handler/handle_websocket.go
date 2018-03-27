package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/websocket"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/message"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/models"
	"github.com/laincloud/entry/server/util"
)

const (
	aliveDecectionInterval = 10 * time.Second
	readBufferSize         = 1024
	writeBufferSize        = 10240 //The write buffer size should be large
	asciiHT                = 9
	asciiCR                = 13
	complementTimeout      = 10 * time.Millisecond
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

type sessionReplay struct {
	typescriptFile *os.File
	timingFile     *os.File
}

type pipe struct {
	requestBuffer  chan []byte
	responseBuffer chan []byte
}

type websocketHandlerFunc func(ctx context.Context, conn *websocket.Conn, r *http.Request, g *global.Global)

// HandleWebsocket handle websocket request
func HandleWebsocket(ctx context.Context, f websocketHandlerFunc, r *http.Request, g *global.Global) middleware.Responder {
	return middleware.ResponderFunc(func(w http.ResponseWriter, _ runtime.Producer) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Upgrade websocket protocol error: %s", err.Error())
			return
		}
		defer conn.Close()

		f(ctx, conn, r, g)
	})
}

func handleRequest(conn *websocket.Conn, s *models.Session, sessionWriter io.WriteCloser, wg *sync.WaitGroup, execID string, msgUnmarshaller util.Unmarshaler, g *global.Global, pipe *pipe) {
	var (
		err   error
		wsMsg []byte
		buf   bytes.Buffer
	)
	time.Sleep(time.Second)
	inMsg := message.RequestMessage{}
	for err == nil {
		if _, wsMsg, err = conn.ReadMessage(); err == nil {
			if unmarshalErr := msgUnmarshaller(wsMsg, &inMsg); unmarshalErr == nil {
				switch inMsg.MsgType {
				case message.RequestMessage_PLAIN:
					if _, err = sessionWriter.Write(inMsg.Content); err != nil {
						log.Errorf("sessionWriter.Write() failed, inMsg.Content: %s(%v), error: %s, session: %+v.", inMsg.Content, inMsg.Content, err, s)
						continue
					}

					switch {
					case len(inMsg.Content) == 1 && inMsg.Content[0] == asciiHT:
						pipe.requestBuffer <- inMsg.Content
						select {
						case complement := <-pipe.responseBuffer:
							buf.Write(complement)
						case <-time.After(complementTimeout):
							log.Error("Command complement timeout, will give up.")
						}
					case len(inMsg.Content) == 1 && inMsg.Content[0] == asciiCR:
						commandContent := string(util.TermEscape(buf.Bytes()))
						if commandContent != "" {
							command := models.Command{
								SessionID: s.SessionID,
								User:      s.User,
								Content:   commandContent,
							}
							g.DB.Create(&command)
							if command.IsRisky() {
								log.Warnf("Dangerous command! Will alert entry owners... Command.Content: %v, session: %+v.", command.Content, s)
								go func() {
									if err1 := command.Alert(*s, g); err1 != nil {
										log.Errorf("command.Alert() failed, error: %v.", err1)
									}
								}()
							} else {
								log.Infof("command.Content: %v, session: %+v.", command.Content, s)
							}
						}
						buf.Reset()
					default:
						if _, err = buf.Write(inMsg.Content); err != nil {
							log.Errorf("buf.Write() failed, error: %s, session: %+v.", err, s)
						} else {
							log.Infof("buf.Write() succeed, inMsg.Content: %s(%v), session: %+v.", inMsg.Content, inMsg.Content, s)
						}
					}
				case message.RequestMessage_WINCH:
					if width, height := util.GetWidthAndHeight(inMsg.Content); width >= 0 && height >= 0 {
						err = g.DockerClient.ResizeExecTTY(execID, height, width)
					}
				}

			} else {
				log.Errorf("Unmarshall request failed, error: %s, session: %+v.", unmarshalErr.Error(), s)
			}
		}
	}
	if err != nil {
		log.Errorf("handleRequest failed, error: %s, session: %+v.", err.Error(), s)
	}

	sessionWriter.Close()
	wg.Done()
}

func handleResponse(conn *websocket.Conn, sessionReader io.ReadCloser, wg *sync.WaitGroup, respType message.ResponseMessage_ResponseType, msgMarshaller util.Marshaler, writeLock *sync.Mutex, replay *sessionReplay, pipe *pipe) {
	var (
		err  error
		size int
	)
	buf := make([]byte, writeBufferSize)
	cursor := 0
	oldTime := time.Now()
	for err == nil {
		if size, err = sessionReader.Read(buf[cursor:]); err == nil || (err == io.EOF && size > 0) {
			validLen := util.GetValidUT8Length(buf[:cursor+size])
			if validLen == 0 {
				log.Errorf("No valid UTF8 sequence prefix")
				break
			}

			if pipe != nil {
				select {
				case <-pipe.requestBuffer:
					log.Infof("buf[:validLen]: %s(%v).", buf[:validLen], buf[:validLen])
					pipe.responseBuffer <- util.TermComplement(buf[:validLen])
				default:
				}
			}

			if replay != nil {
				replay.typescriptFile.Write(buf[:validLen])
				newTime := time.Now()
				delay := newTime.Sub(oldTime)
				oldTime = newTime
				fmt.Fprintf(replay.timingFile, "%f %d\n", float64(delay)/1e9, validLen)
			}

			outMsg := &message.ResponseMessage{
				MsgType: respType,
				Content: buf[:validLen],
			}
			data, marshalErr := msgMarshaller(outMsg)
			if marshalErr == nil {
				writeLock.Lock()
				err = conn.WriteMessage(websocket.BinaryMessage, data)
				writeLock.Unlock()
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
	wg.Done()
}

func handleAliveDetection(ws *websocket.Conn, isStop <-chan int, msgMarshaller util.Marshaler, writeLock *sync.Mutex) {
	pingMsg := &message.ResponseMessage{
		MsgType: message.ResponseMessage_PING,
		Content: []byte("ping"),
	}
	data, _ := msgMarshaller(pingMsg)
	ticker := time.NewTicker(aliveDecectionInterval)
	for {
		select {
		case <-isStop:
			return
		case <-ticker.C:
			writeLock.Lock()
			ws.WriteMessage(websocket.BinaryMessage, data)
			writeLock.Unlock()
		}
	}
}
