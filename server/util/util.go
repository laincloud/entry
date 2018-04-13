package util

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/laincloud/lainlet/message"

	"github.com/laincloud/entry/server/global"
)

const (
	ErrMsgTemplate = "\033[31m>>> %s\033[0m"
)

var (
	errContainerNotfound = errors.New("get data successfully but not found the container")
)

type ViaMethod int

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
func GetContainer(appName, procName, instanceNo string, g *global.Global) (*message.Container, error) {
	coreInfo, err := g.LAINLETClient.CoreinfoGet(appName)
	if err != nil {
		return nil, err
	}

	for procFullName, procInfo := range coreInfo.Data {
		curAppName, curProcName := getAppProcName(strings.Split(procFullName, "."))
		if curProcName == procName && curAppName == appName {
			for _, containerInfo := range procInfo.PodInfos {
				if fmt.Sprint(containerInfo.InstanceNo) == instanceNo &&
					len(containerInfo.Containers) > 0 &&
					containerInfo.Containers[0].Id != "" {
					return containerInfo.Containers[0], nil
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
