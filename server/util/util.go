package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/laincloud/entry/server/global"
)

const (
	ErrMsgTemplate = "\033[31m>>> %s\033[0m"
)

var (
	errContainerNotfound = errors.New("get data successfully but not found the container")
)

type CoreInfo map[string]AppInfo

type ViaMethod int

type Container struct {
	ContainerID string `json:"ContainerId"`
	NodeIP      string `json:"NodeIp"`
}

type AppInfo struct {
	PodInfos []PodInfo `json:"PodInfos"`
}

type PodInfo struct {
	InstanceNo int         `json:"InstanceNo"`
	Containers []Container `json:"ContainerInfos"`
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

// GetContainer get container according to appName, procName and instanceNo
func GetContainer(appName, procName, instanceNo string, g *global.Global) (*Container, error) {
	var (
		data []byte
		err  error
	)
	if data, err = g.LAINLETClient.Get("v2/coreinfowatcher?appname="+appName, 2*time.Second); err != nil {
		return nil, err
	}
	coreInfo := make(CoreInfo)
	if err := json.Unmarshal(data, &coreInfo); err != nil {
		return nil, err
	}
	for procFullName, procInfo := range coreInfo {
		curAppName, curProcName := getAppProcName(strings.Split(procFullName, "."))
		if curProcName == procName && curAppName == appName {
			for _, containerInfo := range procInfo.PodInfos {
				if strconv.Itoa(containerInfo.InstanceNo) == instanceNo &&
					len(containerInfo.Containers) > 0 &&
					containerInfo.Containers[0].ContainerID != "" {
					return &containerInfo.Containers[0], nil
				}
			}
		}
	}
	return nil, errContainerNotfound
}

func GetSourceIP(r *http.Request) string {
	if r.Header.Get("X-Real-IP") != "" {
		return r.Header.Get("X-Real-IP")
	}

	return r.RemoteAddr
}

func GetValidUT8Length(data []byte) int {
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

func GetWidthAndHeight(data []byte) (int, int) {
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
