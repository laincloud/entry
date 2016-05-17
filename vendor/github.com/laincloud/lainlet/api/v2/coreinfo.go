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

type Container struct {
	Command  []string
	Id       string `json:"ContainerId"`
	Ip       string `json:"ContainerIp"`
	Cpu      int
	Env      []string
	Expose   int
	Image    string
	Memory   int64
	NodeIp   string
	NodeName string
	Volumes  []string
}

type Dependency struct {
	PodName string
	Policy  int
}

type PodInfo struct {
	Annotation   string
	Containers   []Container `json:"ContainerInfos"`
	Dependencies []Dependency
	InstanceNo   int
}

// Coreinfo type
type CoreInfo struct {
	PodInfos []PodInfo
}

// Coreinfo API
type GeneralCoreInfo struct {
	Data map[string]CoreInfo
}

func (gci *GeneralCoreInfo) Decode(r []byte) error {
	return json.Unmarshal(r, &gci.Data)
}

func (gci *GeneralCoreInfo) Encode() ([]byte, error) {
	return json.Marshal(gci.Data)
}

func (gci *GeneralCoreInfo) URI() string {
	return "/coreinfowatcher"
}

func (gci *GeneralCoreInfo) WatcherName() string {
	return watcher.PODGROUP
}

func (gci *GeneralCoreInfo) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &GeneralCoreInfo{
		Data: make(map[string]CoreInfo),
	}
	for _, item := range data {
		pg := item.(podgroup.PodGroup)
		ci := CoreInfo{
			PodInfos: make([]PodInfo, len(pg.Pods)),
		}
		for i, pod := range pg.Pods {
			ci.PodInfos[i] = PodInfo{
				Annotation:   pg.Spec.Pod.Annotation,
				InstanceNo:   pod.InstanceNo,
				Containers:   make([]Container, len(pod.Containers)),
				Dependencies: make([]Dependency, len(pg.Spec.Pod.Dependencies)),
			}
			for j, container := range pod.Containers {
				ci.PodInfos[i].Containers[j] = Container{
					Command:  pg.Spec.Pod.Containers[j].Command,
					Id:       container.Id,
					Ip:       container.ContainerIp,
					Cpu:      pg.Spec.Pod.Containers[j].CpuLimit,
					Env:      pg.Spec.Pod.Containers[j].Env,
					Expose:   pg.Spec.Pod.Containers[j].Expose,
					Image:    pg.Spec.Pod.Containers[j].Image,
					Memory:   pg.Spec.Pod.Containers[j].MemoryLimit,
					NodeIp:   container.NodeIp,
					NodeName: container.NodeName,
					Volumes:  pg.Spec.Pod.Containers[j].Volumes,
				}
			}
			for k, depend := range pg.Spec.Pod.Dependencies {
				ci.PodInfos[i].Dependencies[k] = Dependency{
					PodName: depend.PodName,
					Policy:  int(depend.Policy),
				}
			}
		}
		ret.Data[pg.Spec.Name] = ci
	}
	return ret, !reflect.DeepEqual(ret.Data, gci.Data), nil
}

func (gci *GeneralCoreInfo) Key(r *http.Request) (string, error) {
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
