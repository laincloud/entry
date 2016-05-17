package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/podgroup"
)

var hostName, _ = os.Hostname()

type PodInfoForRebellion struct {
	Annotation string
	InstanceNo int
}

// Coreinfo type
type CoreInfoForRebellion struct {
	PodInfos []PodInfoForRebellion
}

// Coreinfo API
type RebellionAPIProvider struct {
	Data map[string]CoreInfoForRebellion
}

func (ap *RebellionAPIProvider) Decode(r []byte) error {
	return json.Unmarshal(r, &ap.Data)
}

func (ap *RebellionAPIProvider) Encode() ([]byte, error) {
	return json.Marshal(ap.Data)
}

func (ap *RebellionAPIProvider) URI() string {
	return "/rebellion/localprocs"
}

func (ap *RebellionAPIProvider) WatcherName() string {
	return watcher.PODGROUP
}

func (ap *RebellionAPIProvider) Make(data map[string]interface{}) (api.API, bool, error) {
	hostName, _ := os.Hostname()
	ret := &RebellionAPIProvider{
		Data: make(map[string]CoreInfoForRebellion),
	}
	for _, item := range data {
		pg := item.(podgroup.PodGroup)
		ci := CoreInfoForRebellion{
			PodInfos: make([]PodInfoForRebellion, 0, len(pg.Pods)),
		}
		for _, pod := range pg.Pods {
			isLocalContainer := false
			for _, container := range pod.Containers {
				if container.NodeName == hostName {
					isLocalContainer = true
					break
				}
			}
			if isLocalContainer {
				ci.PodInfos = append(
					ci.PodInfos,
					PodInfoForRebellion{
						Annotation: pg.Spec.Pod.Annotation,
						InstanceNo: pod.InstanceNo,
					})
			}
		}
		if len(ci.PodInfos) > 0 {
			ret.Data[pg.Spec.Name] = ci
		}
	}
	return ret, !reflect.DeepEqual(ap.Data, ret.Data), nil
}

func (ap *RebellionAPIProvider) Key(r *http.Request) (string, error) {
	appName := api.GetString(r, "appname", "*")
	var err error
	if !auth.Pass(r.RemoteAddr, appName) {
		if appName == "*" { // try to set the appname automatically by remoteip
			appName, err = auth.AppName(r.RemoteAddr)
			if err != nil {
				return "", fmt.Errorf("authorize failed, can not confirm the app by request ip")
			}
			return appName, nil
		}
		return "", fmt.Errorf("authorize failed, no permission")
	}
	return appName, nil
}
