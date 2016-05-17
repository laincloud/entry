package v2

import (
	"encoding/json"
	"fmt"
	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/podgroup"
	"net/http"
	"reflect"
)

// Container info, aim to make it compatible with old api, having to defined a new struct.
type ContainerForProxy struct {
	ContainerIp   string `json:"container_ip"`
	ContainerPort int    `json:"container_port"`
}

type ProcInfo struct {
	Containers []ContainerForProxy `json:"containers"`
}

// Proxy API
type ProxyData struct {
	Data map[string]ProcInfo
}

func (pd *ProxyData) Decode(r []byte) error {
	return json.Unmarshal(r, &pd.Data)
}

func (pd *ProxyData) Encode() ([]byte, error) {
	return json.Marshal(pd.Data)
}

func (pd *ProxyData) URI() string {
	return "/proxywatcher"
}

func (pd *ProxyData) WatcherName() string {
	return watcher.PODGROUP
}

func (pd *ProxyData) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &ProxyData{
		Data: make(map[string]ProcInfo),
	}
	for _, item := range data {
		pg := item.(podgroup.PodGroup)
		pi := ProcInfo{
			Containers: make([]ContainerForProxy, 0),
		}
		for _, pod := range pg.Pods {
			for j, container := range pod.Containers {
				pi.Containers = append(pi.Containers, ContainerForProxy{
					ContainerIp:   container.ContainerIp,
					ContainerPort: pg.Spec.Pod.Containers[j].Expose,
				})
			}

		}
		ret.Data[pg.Spec.Name] = pi
	}
	return ret, !reflect.DeepEqual(pd.Data, ret.Data), nil
}

func (pd *ProxyData) Key(r *http.Request) (string, error) {
	appName := api.GetString(r, "appname", "*")
	if !auth.Pass(r.RemoteAddr, appName) {
		if appName == "*" { // try to set the appname automatically by remoteip
			appName, err := auth.AppName(r.RemoteAddr)
			if err != nil {
				return "", fmt.Errorf("authorize failed, can not confirm the app by request ip")
			}
			return appName, nil
		}
		return "", fmt.Errorf("authorize failed, no permission")
	}
	return appName, nil
}
