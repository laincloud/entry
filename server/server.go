package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/laincloud/entry/message"
	lainlet "github.com/laincloud/lainlet/client"
	"github.com/mijia/sweb/log"
)

type EntryServer struct {
	dockerClient  *docker.Client
	lainletClient *lainlet.Client
	httpClient    *http.Client
}

type ConsoleAuthConf struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ConsoleRole struct {
	Role string `json:"role"`
}

type ConsoleAuthResponse struct {
	Message string      `json:"msg"`
	URL     string      `json:"url"`
	Role    ConsoleRole `json:"role"`
}

type CoreInfo map[string]AppInfo

type Container struct {
	ContainerID string `json:"ContainerId"`
}

type AppInfo struct {
	PodInfos []PodInfo `json:"PodInfos"`
}

type PodInfo struct {
	InstanceNo int         `json:"InstanceNo"`
	Containers []Container `json:"ContainerInfos"`
}

const (
	readBufferSize  = 1024
	writeBufferSize = 10240 //The write buffer size should be large
	byebyeMsg       = "\033[32m>>> You quit the container safely.\033[0m"
	errMsgTemplate  = "\033[31m>>> %s\033[0m"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	errAuthFailed       = fmt.Errorf("authorize failed")
	errAuthNotSupported = fmt.Errorf("entry only works on lain-sso authorization")
	lainDomain          = os.Getenv("LAIN_DOMAIN")
)

//StartServer starts an EntryServer listening on port and connects to DockerSwarm with endpoint.
func StartServer(port, endpoint string) {
	var server *EntryServer
	for {
		if client, err := docker.NewClient(endpoint); err != nil {
			log.Errorf("Initialize docker client error: %s", err.Error())
			time.Sleep(time.Second * 10)
		} else {
			server = &EntryServer{
				dockerClient:  client,
				lainletClient: lainlet.New(net.JoinHostPort("lainlet.lain", os.Getenv("LAINLET_PORT"))),
				httpClient: &http.Client{
					Timeout: 2 * time.Second,
				},
			}
			break
		}
	}

	http.Handle("/enter", *server)
	log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), nil))
}

func (server EntryServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("access-token")
	appName := r.Header.Get("app-name")
	procName := r.Header.Get("proc-name")
	instanceNo := r.Header.Get("instance-no")

	execCmd := []string{"env", fmt.Sprintf("TERM=%s", r.Header.Get("term-type")), "/bin/bash"}

	var (
		err  error
		ws   *websocket.Conn
		exec *docker.Exec
	)

	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Upgrade websocket protocol error: %s", err.Error())
		return
	}
	defer ws.Close()

	if err = server.auth(accessToken, appName); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Authorization failed.")
		log.Errorf("Authorization failed: %s", err.Error())
		server.sendCloseMessage(ws, []byte(errMsg))
		return
	}

	var containerID string
	if containerID, err = server.getContainerID(appName, procName, instanceNo); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Container is not found.")
		log.Errorf("Find container %s[%s-%s] error: %s", appName, procName, instanceNo, err.Error())
		server.sendCloseMessage(ws, []byte(errMsg))
		return
	}

	opts := docker.CreateExecOptions{
		Container:    containerID,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          execCmd,
	}

	if exec, err = server.dockerClient.CreateExec(opts); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Can't enter your container, try again.")
		log.Errorf("Create exec failed: %s", err.Error())
		server.sendCloseMessage(ws, []byte(errMsg))
		return
	}

	stdinPipeReader, stdinPipeWriter := io.Pipe()
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go server.handleRequest(ws, stdinPipeWriter, wg, exec.ID)
	go server.handleResponse(ws, stdoutPipeReader, wg, message.ResponseMessage_STDOUT)
	go server.handleResponse(ws, stderrPipeReader, wg, message.ResponseMessage_STDERR)
	server.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		Detach:       false,
		OutputStream: stdoutPipeWriter,
		ErrorStream:  stderrPipeWriter,
		InputStream:  stdinPipeReader,
		RawTerminal:  false,
	})

	// send close message to client if needed
	server.sendCloseMessage(ws, []byte(byebyeMsg))
	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()
	stdinPipeReader.Close()
	wg.Wait()
}

func (server *EntryServer) handleRequest(ws *websocket.Conn, sessionWriter io.WriteCloser, wg *sync.WaitGroup, execID string) {
	var (
		err   error
		wsMsg []byte
	)
	inMsg := message.RequestMessage{}
	for err == nil {
		if _, wsMsg, err = ws.ReadMessage(); err == nil {
			if unmarshalErr := proto.Unmarshal(wsMsg, &inMsg); unmarshalErr == nil {
				switch inMsg.MsgType {
				case message.RequestMessage_PLAIN:
					if len(inMsg.Content) > 0 {
						_, err = sessionWriter.Write(inMsg.Content)
					}
				case message.RequestMessage_WINCH:
					if width, height := getWidthAndHeight(inMsg.Content); width >= 0 && height >= 0 {
						err = server.dockerClient.ResizeExecTTY(execID, height, width)
					}
				}

			} else {
				log.Errorf("Unmarshall request error: %s", unmarshalErr.Error())
			}
		}
	}
	sessionWriter.Close()
	wg.Done()
}

func (server *EntryServer) handleResponse(ws *websocket.Conn, sessionReader io.ReadCloser, wg *sync.WaitGroup, respType message.ResponseMessage_ResponseType) {
	var (
		err  error
		size int
	)
	buf := make([]byte, writeBufferSize)
	cursor := 0
	for err == nil {
		if size, err = sessionReader.Read(buf[cursor:]); err == nil || (err == io.EOF && size > 0) {
			validLen := getValidUT8Length(buf[:cursor+size])
			if validLen == 0 {
				log.Errorf("No valid UTF8 sequence prefix")
				break
			}
			outMsg := &message.ResponseMessage{
				MsgType: respType,
				Content: buf[:validLen],
			}
			data, marshalErr := proto.Marshal(outMsg)
			if marshalErr == nil {
				err = ws.WriteMessage(websocket.BinaryMessage, data)
				cursor := size - validLen
				for i := 0; i < cursor; i++ {
					buf[i] = buf[cursor+i]
				}
			} else {
				log.Errorf("Marshal response error: %s", marshalErr.Error())
			}
		}
	}
	sessionReader.Close()
	wg.Done()
}

// auth authorizes whether the client with the token has the right to access the application
func (server *EntryServer) auth(token, appName string) error {
	var (
		data []byte
		err  error
	)
	if data, err = server.lainletClient.Get("/v2/configwatcher?target=auth/console", 2*time.Second); err != nil {
		return err
	}
	authDataMap := make(map[string]string)
	if err = json.Unmarshal(data, &authDataMap); err != nil {
		return err
	}
	if authStr, exist := authDataMap["auth/console"]; exist {
		c := ConsoleAuthConf{}
		if err = json.Unmarshal([]byte(authStr), &c); err != nil {
			return err
		}
		if c.Type == "lain-sso" {
			authURL := fmt.Sprintf("http://console.%s/api/v1/repos/%s/roles/", lainDomain, appName)
			return server.validateConsoleRole(authURL, token)
		}
		return errAuthNotSupported
	}

	return nil
}

func (server *EntryServer) validateConsoleRole(authURL, token string) error {
	var (
		err       error
		req       *http.Request
		resp      *http.Response
		respBytes []byte
	)
	if req, err = http.NewRequest("GET", authURL, nil); err != nil {
		return err
	}
	req.Header.Set("access-token", token)
	if resp, err = server.httpClient.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()
	if respBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}
	caResp := ConsoleAuthResponse{}
	if err = json.Unmarshal(respBytes, &caResp); err != nil {
		return err
	}
	if caResp.Role.Role == "" {
		return errAuthFailed
	}
	return nil
}

func (server *EntryServer) getContainerID(appName, procName, instanceNo string) (string, error) {
	var (
		data []byte
		err  error
	)
	if data, err = server.lainletClient.Get("v2/coreinfowatcher?appname="+appName, 2*time.Second); err != nil {
		return "", err
	}
	coreInfo := make(CoreInfo)
	if err := json.Unmarshal(data, &coreInfo); err != nil {
		return "", err
	}
	for procFullName, procInfo := range coreInfo {
		keyParts := strings.Split(procFullName, ".")
		if len(keyParts) > 0 && keyParts[len(keyParts)-1] == procName {
			for _, containerInfo := range procInfo.PodInfos {
				if strconv.Itoa(containerInfo.InstanceNo) == instanceNo &&
					len(containerInfo.Containers) > 0 &&
					containerInfo.Containers[0].ContainerID != "" {
					return containerInfo.Containers[0].ContainerID, nil
				}
			}
		}
	}
	return "", fmt.Errorf("get data successfully but not found the container")
}

func (server *EntryServer) sendCloseMessage(ws *websocket.Conn, content []byte) {
	closeMsg := &message.ResponseMessage{
		MsgType: message.ResponseMessage_CLOSE,
		Content: content,
	}
	if closeData, err := proto.Marshal(closeMsg); err != nil {
		log.Errorf("Marshal close message failed: %s", err.Error())
	} else {
		ws.WriteMessage(websocket.BinaryMessage, closeData)
	}
}

func getWidthAndHeight(data []byte) (int, int) {
	sizeStr := string(data)
	sizeArr := strings.Split(sizeStr, " ")

	if len(sizeArr) != 2 {
		return -1, -1
	}
	var width, height int
	var err error

	if width, err = strconv.Atoi(sizeArr[0]); err != nil {
		return -1, -1
	}
	if height, err = strconv.Atoi(sizeArr[1]); err != nil {
		return -1, -1
	}

	return width, height
}

func getValidUT8Length(data []byte) int {
	validLen := 0
	for i := len(data) - 1; i >= 0; i-- {
		if utf8.RuneStart(data[i]) {
			validLen = i
			if utf8.Valid(data[i:]) {
				validLen = len(data)
			}
			break
		}
	}
	return validLen
}
