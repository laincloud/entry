package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

const (
	byebyeMsg = "\033[32m>>> You quit the container safely.\033[0m"
)

// Enter enter to container
func Enter(ctx context.Context, conn *websocket.Conn, r *http.Request, g *global.Global) {
	s, err := models.NewSession(conn, r, g)
	if err != nil {
		log.Errorf("models.NewSession() failed, error: %s.", err)
		return
	}

	g.DB.Create(s)
	defer func() {
		g.DB.Model(s).Updates(models.Session{
			Status:  models.SessionStatusInactive,
			EndedAt: time.Now(),
		})
	}()

	if err := os.MkdirAll(s.DataPath(), 0700); err != nil {
		log.Errorf("os.MkdirAll(%s) failed, error: %s.", s.DataPath(), err)
		return
	}

	sessionReplay, err := pipe.NewSessionReplay(*s)
	if err != nil {
		log.Errorf("pipe.NewSessionReplay(%v) failed, error: %s.", s, err)
		return
	}
	defer sessionReplay.Close()

	termType := r.Header.Get("term-type")
	if len(termType) == 0 {
		termType = "xterm-256color"
	}

	execCmd := []string{"env", fmt.Sprintf("TERM=%s", termType), "/bin/bash"}
	opts := docker.CreateExecOptions{
		Container:    s.ContainerID,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          execCmd,
	}

	msgMarshaller, msgUnmarshaller := util.GetMarshalers(r)
	writeLock := &sync.Mutex{}

	exec, err := g.DockerClient.CreateExec(opts)
	if err != nil {
		errMsg := fmt.Sprintf(util.ErrMsgTemplate, "Can't enter your container, try again.")
		log.Errorf("Create exec failed, error: %s, session: %+v.", err.Error(), s)
		util.SendCloseMessage(conn, []byte(errMsg), msgMarshaller, writeLock)
		return
	}

	stdinPipeReader, stdinPipeWriter := io.Pipe()
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	stopSignal := make(chan int)
	wg := &sync.WaitGroup{}
	p := pipe.NewPipe(conn, msgMarshaller, s, msgUnmarshaller, wg, writeLock)
	wg.Add(3)
	go p.HandleAliveDetection(stopSignal)
	go p.HandleRequest(exec.ID, stdinPipeWriter, g)
	go p.HandleResponse(message.ResponseMessage_STDOUT, stdoutPipeReader, sessionReplay)
	go p.HandleResponse(message.ResponseMessage_STDERR, stderrPipeReader, sessionReplay)
	go func() {
		if err = g.DockerClient.StartExec(exec.ID, docker.StartExecOptions{
			Detach:       false,
			OutputStream: stdoutPipeWriter,
			ErrorStream:  stderrPipeWriter,
			InputStream:  stdinPipeReader,
			RawTerminal:  false,
		}); err != nil {
			errMsg := fmt.Sprintf(util.ErrMsgTemplate, "Can't enter your container, try again.")
			log.Errorf("Start exec failed, error: %s, session: %+v.", err.Error(), s)
			util.SendCloseMessage(conn, []byte(errMsg), msgMarshaller, writeLock)
		} else {
			util.SendCloseMessage(conn, []byte(byebyeMsg), msgMarshaller, writeLock)
		}
		close(stopSignal)
	}()

	select {
	case <-ctx.Done():
		log.Infof("Entering to %s canceled, session: %+v.", s.ContainerID, s)
	case <-stopSignal:
		log.Infof("Entering to %s stopped, session: %+v", s.ContainerID, s)
	}
	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()
	stdinPipeReader.Close()
	wg.Wait()
}
