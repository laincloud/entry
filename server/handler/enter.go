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
	"github.com/laincloud/entry/message"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/models"
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

	typescriptFile, err := os.Create(s.TypescriptFile())
	if err != nil {
		log.Errorf("os.Create(%s) failed, error: %s.", s.TypescriptFile(), err)
		return
	}
	fmt.Fprintf(typescriptFile, "Script started on %s\n", time.Now())
	defer func() {
		fmt.Fprintf(typescriptFile, "Script done on %s\n", time.Now())
		typescriptFile.Close()
	}()

	timingFile, err := os.Create(s.TimingFile())
	if err != nil {
		log.Errorf("os.Create(%s) failed, error: %s.", s.TimingFile(), err)
		return
	}
	defer timingFile.Close()

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
	wg.Add(3)
	go handleAliveDetection(conn, stopSignal, msgMarshaller, writeLock)
	go handleRequest(conn, s, stdinPipeWriter, wg, exec.ID, msgUnmarshaller, g)
	go handleResponse(conn, stdoutPipeReader, wg, message.ResponseMessage_STDOUT, msgMarshaller, writeLock, typescriptFile, timingFile)
	go handleResponse(conn, stderrPipeReader, wg, message.ResponseMessage_STDERR, msgMarshaller, writeLock, typescriptFile, timingFile)
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
