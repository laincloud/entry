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
type ViaMethod int
type Marshaler func(interface{}) ([]byte, error)
type Unmarshaler func([]byte, interface{}) error

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

	http.HandleFunc("/enter", server.enter)
	http.HandleFunc("/attach", server.attach)
	log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), nil))
}

func (server *EntryServer) enter(w http.ResponseWriter, r *http.Request) {
	ws, containerID, err := server.prepare(w, r)
	if ws != nil {
		defer ws.Close()
	}
	if err != nil {
		return
	}
	var exec *docker.Exec

	termType := r.Header.Get("term-type")
	if len(termType) == 0 {
		termType = "xterm-256color"
	}

	execCmd := []string{"env", fmt.Sprintf("TERM=%s", termType), "/bin/bash"}
	opts := docker.CreateExecOptions{
		Container:    containerID,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          execCmd,
	}

	msgMarshaller, msgUnmarshaller := getMarshalers(r)

	if exec, err = server.dockerClient.CreateExec(opts); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Can't enter your container, try again.")
		log.Errorf("Create exec failed: %s", err.Error())
		server.sendCloseMessage(ws, []byte(errMsg), msgMarshaller)
		return
	}

	stdinPipeReader, stdinPipeWriter := io.Pipe()
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go server.handleRequest(ws, stdinPipeWriter, wg, exec.ID, msgUnmarshaller)
	go server.handleResponse(ws, stdoutPipeReader, wg, message.ResponseMessage_STDOUT, msgMarshaller)
	go server.handleResponse(ws, stderrPipeReader, wg, message.ResponseMessage_STDERR, msgMarshaller)
	if err = server.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		Detach:       false,
		OutputStream: stdoutPipeWriter,
		ErrorStream:  stderrPipeWriter,
		InputStream:  stdinPipeReader,
		RawTerminal:  false,
	}); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Can't enter your container, try again.")
		log.Errorf("Start exec failed: %s", err.Error())
		server.sendCloseMessage(ws, []byte(errMsg), msgMarshaller)
	} else {
		server.sendCloseMessage(ws, []byte(byebyeMsg), msgMarshaller)
	}

	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()
	stdinPipeReader.Close()
	wg.Wait()
	log.Infof("Entering to %s stopped", containerID)
}

func (server *EntryServer) attach(w http.ResponseWriter, r *http.Request) {
	ws, containerID, err := server.prepare(w, r)
	if ws != nil {
		defer ws.Close()
	}
	if err != nil {
		return
	}
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(2)

	opts := docker.AttachToContainerOptions{
		Container:    containerID,
		Stdin:        false,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		OutputStream: stdoutPipeWriter,
		ErrorStream:  stderrPipeWriter,
	}

	msgMarshaller, _ := getMarshalers(r)
	go server.handleResponse(ws, stdoutPipeReader, wg, message.ResponseMessage_STDOUT, msgMarshaller)
	go server.handleResponse(ws, stderrPipeReader, wg, message.ResponseMessage_STDERR, msgMarshaller)

	if waiter, err := server.dockerClient.AttachToContainerNonBlocking(opts); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Can't attach your container, try again.")
		log.Errorf("Attach failed: %s", err.Error())
		server.sendCloseMessage(ws, []byte(errMsg), msgMarshaller)
	} else {
		// Check whether the websocket is closed
		for {
			if _, _, err = ws.ReadMessage(); err == nil {
				time.Sleep(10 * time.Millisecond)
			} else {
				break
			}
		}
		waiter.Close()
	}
	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()
	wg.Wait()
	log.Infof("Attaching to %s stopped", containerID)
}

func (server *EntryServer) prepare(w http.ResponseWriter, r *http.Request) (*websocket.Conn, string, error) {
	accessToken := r.Header.Get("access-token")
	msgMarshaller, _ := getMarshalers(r)

	appName := r.Header.Get("app-name")
	procName := r.Header.Get("proc-name")
	instanceNo := r.Header.Get("instance-no")

	var containerID string
	log.Infof("A user wants to enter %s[%s-%s]", appName, procName, instanceNo)

	var (
		err error
		ws  *websocket.Conn
	)

	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Upgrade websocket protocol error: %s", err.Error())
		return ws, containerID, err
	}

	if err = server.auth(accessToken, appName); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Authorization failed.")
		log.Errorf("Authorization failed: %s", err.Error())
		server.sendCloseMessage(ws, []byte(errMsg), msgMarshaller)
		return ws, containerID, err
	}

	if containerID, err = server.getContainerID(appName, procName, instanceNo); err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Container is not found.")
		log.Errorf("Find container %s[%s-%s] error: %s", appName, procName, instanceNo, err.Error())
		server.sendCloseMessage(ws, []byte(errMsg), msgMarshaller)
	}
	return ws, containerID, err
}

func (server *EntryServer) handleRequest(ws *websocket.Conn, sessionWriter io.WriteCloser, wg *sync.WaitGroup, execID string, msgUnmarshaller Unmarshaler) {
	var (
		err   error
		wsMsg []byte
	)
	time.Sleep(time.Second)
	inMsg := message.RequestMessage{}
	for err == nil {
		if _, wsMsg, err = ws.ReadMessage(); err == nil {
			if unmarshalErr := msgUnmarshaller(wsMsg, &inMsg); unmarshalErr == nil {
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
	if err != nil {
		log.Errorf("HandleRequest ended: %s", err.Error())
	}

	sessionWriter.Close()
	wg.Done()
}

func (server *EntryServer) handleResponse(ws *websocket.Conn, sessionReader io.ReadCloser, wg *sync.WaitGroup, respType message.ResponseMessage_ResponseType, msgMarshaller Marshaler) {
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
			data, marshalErr := msgMarshaller(outMsg)
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
	if err != nil {
		log.Errorf("HandleResponse ended: %s", err.Error())
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
		curAppName, curProcName := getAppProcName(strings.Split(procFullName, "."))
		if curProcName == procName && curAppName == appName {
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

func (server *EntryServer) sendCloseMessage(ws *websocket.Conn, content []byte, msgMarshaller Marshaler) {
	closeMsg := &message.ResponseMessage{
		MsgType: message.ResponseMessage_CLOSE,
		Content: content,
	}
	if closeData, err := msgMarshaller(closeMsg); err != nil {
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

func getAppProcName(key []string) (string, string) {
	var procName string
	if len(key) > 0 {
		procName = key[len(key)-1]
	}
	var tmp []string
	for i := len(key) - 3; i >= 0; i-- {
		tmp = append(tmp, key[i])
	}
	return strings.Join(tmp, "."), procName
}

func getMarshalers(r *http.Request) (Marshaler, Unmarshaler) {
	if r.URL.Query().Get("method") == "web" {
		return json.Marshal, json.Unmarshal
	}
	return protoMarshalFunc, protoUnmarshalFunc
}

// Adapters
func protoMarshalFunc(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func protoUnmarshalFunc(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}
