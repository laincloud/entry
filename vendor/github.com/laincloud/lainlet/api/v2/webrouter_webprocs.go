package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/podgroup"
)

type ContainerForWebrouter struct {
	IP     string `json:"ContainerIp"`
	Expose int
}

type PodInfoForWebrouter struct {
	Annotation string
	Containers []ContainerForWebrouter `json:"ContainerInfos"`
}

// Coreinfo type
type CoreInfoForWebrouter struct {
	PodInfos []PodInfoForWebrouter
}

// Coreinfo API
type WebrouterInfo struct {
	Data map[string]CoreInfoForWebrouter
}

func (wi *WebrouterInfo) Decode(r []byte) error {
	return json.Unmarshal(r, &wi.Data)
}

func (wi *WebrouterInfo) Encode() ([]byte, error) {
	return json.Marshal(wi.Data)
}

func (wi *WebrouterInfo) URI() string {
	return "/webrouter/webprocs"
}

func (wi *WebrouterInfo) WatcherName() string {
	return watcher.PODGROUP
}

func (wi *WebrouterInfo) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &WebrouterInfo{
		Data: make(map[string]CoreInfoForWebrouter),
	}
	for _, item := range data {
		pg := item.(podgroup.PodGroup)
		parts := strings.Split(pg.Spec.Name, ".")
		// Webrouter only cares about web procs
		if len(parts) < 3 || parts[len(parts)-2] != "web" {
			continue
		}
		ci := CoreInfoForWebrouter{
			PodInfos: make([]PodInfoForWebrouter, len(pg.Pods)),
		}
		for i, pod := range pg.Pods {
			ci.PodInfos[i] = PodInfoForWebrouter{
				Annotation: pg.Spec.Pod.Annotation,
				Containers: make([]ContainerForWebrouter, len(pod.Containers)),
			}
			for j, container := range pod.Containers {
				ci.PodInfos[i].Containers[j] = ContainerForWebrouter{
					IP:     container.ContainerIp,
					Expose: pg.Spec.Pod.Containers[j].Expose,
				}
			}
		}
		ret.Data[pg.Spec.Name] = ci
	}
	return ret, !reflect.DeepEqual(wi.Data, ret.Data), nil
}

func (wi *WebrouterInfo) Key(r *http.Request) (string, error) {
	appName := api.GetString(r, "appname", "*")
	if !auth.Pass(r.RemoteAddr, appName) {
		if appName == "*" { // try to set the appname automatically by remoteip
			appName, err := auth.AppName(r.RemoteAddr)
			if err != nil {
				return "", fmt.Errorf("authorize failed, can not confirm the app by request ip")
			}
			return fixPrefix(appName), nil
		}
		return "", fmt.Errorf("authorize failed, no permission")
	}
	if appName != "*" {
		appName = fixPrefix(appName)
	}
	return appName, nil
}
