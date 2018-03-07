package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/websocket"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/message"
	swaggermodels "github.com/laincloud/entry/server/gen/models"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/models"
	"github.com/laincloud/entry/server/util"
)

const (
	aliveDecectionInterval = 10 * time.Second
	readBufferSize         = 1024
	writeBufferSize        = 10240 //The write buffer size should be large
	asciiCR                = 13
	keyAccessToken         = "access_token"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	noAuthAPIPaths = []string{
		"/enter",
		"/attach",
		"/api/authorize",
		"/api/config",
		"/api/logout",
		"/api/ping",
	}
)

type websocketHandlerFunc func(ctx context.Context, conn *websocket.Conn, r *http.Request, s *models.Session, g *global.Global)

// AuthAPI judge whether the request has right to our API
func AuthAPI(h http.Handler, g *global.Global) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("request: %+v.", r)

		for _, p := range noAuthAPIPaths {
			if r.URL.Path == p {
				h.ServeHTTP(w, r)
				return
			}
		}

		accessToken, err := r.Cookie(keyAccessToken)
		if err != nil {
			errMsg := err.Error()
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(swaggermodels.Error{
				Message: &errMsg,
			})
			return
		}

		if _, err = util.AuthAPI(accessToken.Value, g); err != nil {
			errMsg := err.Error()
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(swaggermodels.Error{
				Message: &errMsg,
			})
			return
		}

		h.ServeHTTP(w, r)
	})
}

// HandleWebsocket handle websocket request
func HandleWebsocket(ctx context.Context, sessionType string, f websocketHandlerFunc, r *http.Request, g *global.Global) middleware.Responder {
	return middleware.ResponderFunc(func(w http.ResponseWriter, _ runtime.Producer) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Upgrade websocket protocol error: %s", err.Error())
			return
		}
		defer conn.Close()

		session, err := models.NewSession(sessionType, conn, r, g)
		if err != nil {
			log.Errorf("models.NewSession() failed, error: %s.", err)
			return
		}

		g.DB.Create(session)
		f(ctx, conn, r, session, g)
		g.DB.Model(session).Updates(models.Session{
			Status:  models.SessionStatusInactive,
			EndedAt: time.Now(),
		})
	})
}

func handleRequest(conn *websocket.Conn, s *models.Session, sessionWriter io.WriteCloser, wg *sync.WaitGroup, execID string, msgUnmarshaller util.Unmarshaler, g *global.Global) {
	var (
		n     int
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
					if n, err = sessionWriter.Write(inMsg.Content); err != nil {
						log.Errorf("sessionWriter.Write() failed, inMsg.Content: %s(%v), error: %s, session: %+v.", inMsg.Content, inMsg.Content, err, s)
					} else {
						log.Infof("sessionWriter.Write() succeed, inMsg.Content: %s(%v), n: %d, session: %+v.", inMsg.Content, inMsg.Content, n, s)
					}

					if _, err = buf.Write(inMsg.Content); err != nil {
						log.Errorf("buf.Write() failed, error: %s, session: %+v.", err, s)
					}

					if len(inMsg.Content) == 1 && inMsg.Content[0] == asciiCR {
						command := models.Command{
							SessionID: s.SessionID,
							Session:   *s,
							User:      s.User,
							Content:   buf.String(),
						}
						buf.Reset()
						g.DB.Create(&command)
						log.Infof("command.Content: %s(%v), session: %+v.", command.Content, command.Content, s)
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

func handleResponse(ws *websocket.Conn, sessionReader io.ReadCloser, wg *sync.WaitGroup, respType message.ResponseMessage_ResponseType, msgMarshaller util.Marshaler, writeLock *sync.Mutex) {
	var (
		err  error
		size int
	)
	buf := make([]byte, writeBufferSize)
	cursor := 0
	for err == nil {
		if size, err = sessionReader.Read(buf[cursor:]); err == nil || (err == io.EOF && size > 0) {
			validLen := util.GetValidUT8Length(buf[:cursor+size])
			if validLen == 0 {
				log.Errorf("No valid UTF8 sequence prefix")
				break
			}
			outMsg := &message.ResponseMessage{
				MsgType: respType,
				Content: buf[:validLen],
			}
			data, marshalErr := msgMarshaller(outMsg)
			if marshalErr == nil {
				writeLock.Lock()
				err = ws.WriteMessage(websocket.BinaryMessage, data)
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
