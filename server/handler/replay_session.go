package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/message"
	"github.com/laincloud/entry/server/models"
	"github.com/laincloud/entry/server/util"
)

const replaySessionDoneMsg = "\033[32m>>> Session replay done.\033[0m"

// ReplaySession replay the session
func ReplaySession(ctx context.Context, conn *websocket.Conn, r *http.Request, g *global.Global) {
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) != 5 {
		log.Errorf("r.URL.Path: %s is invalid.", r.URL.Path)
		return
	}

	rawSessionID := paths[3]
	sessionID, err := strconv.ParseInt(rawSessionID, 10, 64)
	if err != nil {
		log.Errorf("strconv.ParseInt(%s, 10, 64) failed, error: %s.", rawSessionID, err)
		return
	}

	var s models.Session
	if err = g.DB.Where("session_id = ?", sessionID).First(&s).Error; err != nil {
		log.Errorf("g.DB.Where(session_id = %d) failed, error: %s.", sessionID, err)
		return
	}

	cmd := exec.Command("scriptreplay", "--timing", s.TimingFile(), s.TypescriptFile())
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("cmd.StdoutPipe() failed, error: %s.", err)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("cmd.StderrPipe() failed, error: %s.", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	msgMarshaller := json.Marshal
	writeLock := &sync.Mutex{}
	stopSignal := make(chan int)
	go handleAliveDetection(conn, stopSignal, msgMarshaller, writeLock)
	go handleResponse(conn, stdoutPipe, wg, message.ResponseMessage_STDOUT, msgMarshaller, writeLock, nil, nil)
	go handleResponse(conn, stderrPipe, wg, message.ResponseMessage_STDERR, msgMarshaller, writeLock, nil, nil)

	go func() {
		if err1 := cmd.Run(); err1 != nil {
			errMsg := fmt.Sprintf(util.ErrMsgTemplate, "Replay session failed, please try again.")
			log.Errorf("Replay session: %+v failed, error: %s.", s, err1)
			util.SendCloseMessage(conn, []byte(errMsg), msgMarshaller, writeLock)
		} else {
			util.SendCloseMessage(conn, []byte(replaySessionDoneMsg), msgMarshaller, writeLock)
		}
		close(stopSignal)
	}()

	select {
	case <-ctx.Done():
		log.Infof("Replay session: %+v canceled.", s)
	case <-stopSignal:
		log.Infof("Replay session: %+v done.", s)
	}
	wg.Wait()
}
