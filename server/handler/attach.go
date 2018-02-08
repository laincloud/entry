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
	"github.com/laincloud/entry/message"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/models"
	"github.com/laincloud/entry/server/util"
)

// Attach attach to container to view stdout/stderr of the container
func Attach(ctx context.Context, conn *websocket.Conn, r *http.Request, session *models.Session, g *global.Global) {
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(2)

	opts := docker.AttachToContainerOptions{
		Container:    session.ContainerID,
		Stdin:        false,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		OutputStream: stdoutPipeWriter,
		ErrorStream:  stderrPipeWriter,
	}

	msgMarshaller, _ := util.GetMarshalers(r)
	writeLock := &sync.Mutex{}
	go handleResponse(conn, stdoutPipeReader, wg, message.ResponseMessage_STDOUT, msgMarshaller, writeLock)
	go handleResponse(conn, stderrPipeReader, wg, message.ResponseMessage_STDERR, msgMarshaller, writeLock)

	stopSignal := make(chan int)
	go func() {
		if waiter, err := g.DockerClient.AttachToContainerNonBlocking(opts); err != nil {
			errMsg := fmt.Sprintf(util.ErrMsgTemplate, "Can't attach your container, try again.")
			log.Errorf("Attach failed: %s", err.Error())
			util.SendCloseMessage(conn, []byte(errMsg), msgMarshaller, writeLock)
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
		log.Infof("Attaching to %s canceled.", session.ContainerID)
	case <-stopSignal:
		log.Infof("Attaching to %s stopped.", session.ContainerID)
	}
	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()
	wg.Wait()
}
