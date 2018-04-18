package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/message"
	"github.com/laincloud/entry/server/models"
	"github.com/laincloud/entry/server/pipe"
	"github.com/laincloud/entry/server/util"
)

// Attach attach to container to view stdout/stderr of the container
func Attach(ctx context.Context, conn *websocket.Conn, r *http.Request, g *global.Global) {
	s, err := models.NewSession(conn, r, g)
	if err != nil {
		log.Errorf("models.NewSession() failed, error: %s.", err)
		return
	}

	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(2)

	opts := docker.AttachToContainerOptions{
		Container:    s.ContainerID,
		Stdin:        false,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		OutputStream: stdoutPipeWriter,
		ErrorStream:  stderrPipeWriter,
	}

	msgMarshaler, msgUnMarshaler := util.GetMarshalers(r)
	writeLock := &sync.Mutex{}
	p := pipe.NewPipe(conn, msgMarshaler, s, msgUnMarshaler, wg, writeLock)
	go p.HandleResponse(message.ResponseMessage_STDOUT, stdoutPipeReader, nil)
	go p.HandleResponse(message.ResponseMessage_STDERR, stderrPipeReader, nil)

	stopSignal := make(chan int)
	go func() {
		if waiter, err := g.DockerClient.AttachToContainerNonBlocking(opts); err != nil {
			errMsg := fmt.Sprintf(util.ErrMsgTemplate, "Can't attach your container, try again.")
			log.Errorf("Attach failed: %s", err.Error())
			util.SendCloseMessage(conn, []byte(errMsg), msgMarshaler, writeLock)
		} else {
			// Check whether the websocket is closed
			for {
				if _, _, err = conn.ReadMessage(); err == nil {
					time.Sleep(10 * time.Millisecond)
				} else {
					break
				}
			}
			waiter.Close()
		}
		close(stopSignal)
	}()

	select {
	case <-ctx.Done():
		log.Infof("Attaching to %s canceled.", s.ContainerID)
	case <-stopSignal:
		log.Infof("Attaching to %s stopped.", s.ContainerID)
	}
	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()
	wg.Wait()
}
