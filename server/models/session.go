package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	swaggermodels "github.com/laincloud/entry/server/gen/models"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/util"
)

const (
	entryAppName          = "entry"
	SessionStatusActive   = "active"
	SessionStatusInactive = "inactive"
	dataPath              = "/cloud/data/sessions"
)

// Session denotes a user session connected to a container
type Session struct {
	SessionID   int64  `gorm:"primary_key"`
	User        string `gorm:"index"`
	SourceIP    string `gorm:"index"`
	AppName     string `gorm:"index"`
	ProcName    string
	InstanceNo  string
	ContainerID string
	NodeIP      string
	Status      string
	CreatedAt   time.Time `sql:"not null;DEFAULT:current_timestamp"`
	EndedAt     time.Time
	UpdatedAt   time.Time `sql:"not null;DEFAULT:current_timestamp"`
}

// NewSession initialize a session
func NewSession(conn *websocket.Conn, r *http.Request, g *global.Global) (*Session, error) {
	isViaWeb := r.URL.Query().Get("method") == "web"
	var accessToken, appName, procName, instanceNo string
	msgMarshaller, _ := util.GetMarshalers(r)
	if !isViaWeb {
		accessToken = r.Header.Get("access-token")
		appName = r.Header.Get("app-name")
		procName = r.Header.Get("proc-name")
		instanceNo = r.Header.Get("instance-no")
	} else {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Read auth message from webclient failed: %s", err.Error())
			return nil, err
		}

		msg := make(map[string]string)
		json.Unmarshal(msgData, &msg)
		accessToken = msg["access_token"]
		appName = msg["app_name"]
		procName = msg["proc_name"]
		instanceNo = msg["instance_no"]
	}

	if appName == entryAppName {
		log.Errorf("appName == %s is not allowed.", entryAppName)
		return nil, util.ErrAuthFailed
	}

	log.Infof("A user wants to enter %s[%s-%s]", appName, procName, instanceNo)
	writeLock := &sync.Mutex{}
	ssoUser, err := util.AuthContainer(accessToken, appName, g)
	if err != nil {
		errMsg := fmt.Sprintf(util.ErrMsgTemplate, "Authorization failed.")
		log.Errorf("Authorization failed: %s", err.Error())
		util.SendCloseMessage(conn, []byte(errMsg), msgMarshaller, writeLock)
		return nil, err
	}

	container, err := util.GetContainer(appName, procName, instanceNo, g)
	if err != nil {
		errMsg := fmt.Sprintf(util.ErrMsgTemplate, "Container is not found.")
		log.Errorf("Find container %s[%s-%s] error: %s", appName, procName, instanceNo, err.Error())
		util.SendCloseMessage(conn, []byte(errMsg), msgMarshaller, writeLock)
		return nil, err
	}

	s := Session{
		User:        ssoUser.Email,
		SourceIP:    util.GetSourceIP(r),
		AppName:     appName,
		ProcName:    procName,
		InstanceNo:  instanceNo,
		ContainerID: container.ContainerID,
		NodeIP:      container.NodeIP,
		Status:      SessionStatusActive,
	}
	log.Infof("A new session: %+v has been created.", s)
	return &s, nil
}

// SwaggerModel return the swagger version
func (s Session) SwaggerModel() swaggermodels.Session {
	return swaggermodels.Session{
		SessionID:   s.SessionID,
		User:        s.User,
		SourceIP:    s.SourceIP,
		AppName:     s.AppName,
		ProcName:    s.ProcName,
		InstanceNo:  s.InstanceNo,
		ContainerID: s.ContainerID,
		NodeIP:      s.NodeIP,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt.Unix(),
		EndedAt:     s.EndedAt.Unix(),
	}
}

// DataPath return the parent directory of typescript file and timing file
func (s Session) DataPath() string {
	return fmt.Sprintf("%s/%d", dataPath, s.SessionID)
}

// TypescriptFile return the file path which record stdout/stderr for scriptrelpay
func (s Session) TypescriptFile() string {
	return fmt.Sprintf("%s/typescript", s.DataPath())
}

// TimingFile return the file path which record time info for scriptrelpay
func (s Session) TimingFile() string {
	return fmt.Sprintf("%s/timing.txt", s.DataPath())
}
