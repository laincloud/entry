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
	"strings"
)

type Pod struct {
	InstanceNo int
	IP         string
	Port       int
	ProcName   string
}

type PodGroup struct {
	Pods []Pod `json:"proc"`
}

// PodGroup API
type GeneralPodGroup struct {
	Data []PodGroup
}

func (gpg *GeneralPodGroup) Decode(r []byte) error {
	return json.Unmarshal(r, &gpg.Data)
}

func (gpg *GeneralPodGroup) Encode() ([]byte, error) {
	return json.Marshal(gpg.Data)
}

func (gpg *GeneralPodGroup) URI() string {
	return "/procwatcher"
}

func (gpg *GeneralPodGroup) WatcherName() string {
	return watcher.PODGROUP
}

func (gpg *GeneralPodGroup) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &GeneralPodGroup{
		Data: make([]PodGroup, 0, len(data)),
	}
	for _, procInfo := range data {
		pg := &PodGroup{
			Pods: make([]Pod, 0),
		}
		for _, instanceInfo := range procInfo.(podgroup.PodGroup).Pods {
			envlist := instanceInfo.Containers[0].Runtime.Config.Env
			procname := ""
			for _, str := range envlist {
				if strings.HasPrefix(str, "LAIN_PROCNAME=") {
					procname = str[len("LAIN_PROCNAME="):]
				}
			}
			inst := &Pod{
				ProcName:   procname,
				InstanceNo: instanceInfo.InstanceNo,
				IP:         instanceInfo.Containers[0].ContainerIp,
				Port:       instanceInfo.Containers[0].ContainerPort,
			}
			pg.Pods = append(pg.Pods, *inst)
		}
		ret.Data = append(ret.Data, *pg)
	}
	return ret, !reflect.DeepEqual(ret.Data, gpg.Data), nil
}

func (gpg *GeneralPodGroup) Key(r *http.Request) (string, error) {
	appName := api.GetString(r, "appname", "")
	if appName == "" {
		return "", fmt.Errorf("appname required")
	}
	if !auth.Pass(r.RemoteAddr, appName) {
		return "", fmt.Errorf("authorize failed, no permission")
	}
	return appName, nil
}
