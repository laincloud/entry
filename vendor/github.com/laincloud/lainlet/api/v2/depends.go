package v2

import (
	"encoding/json"
	"fmt"
	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/depends"
	"net/http"
	"reflect"
)

// The ContainerInfo used for dependency
type ContainerInfo struct {
	ContainerID string
	NodeIP      string
	IP          string
	Port        int
}

type DependsItem struct {
	Annotation string
	Containers []ContainerInfo
}

// Depends API
type Depends struct {
	Data map[string]map[string]map[string]DependsItem
}

func (d *Depends) Decode(r []byte) error {
	return json.Unmarshal(r, &d.Data)
}

func (d *Depends) Encode() ([]byte, error) {
	return json.Marshal(d.Data)
}

func (d *Depends) URI() string {
	return "/depends"
}

func (d *Depends) WatcherName() string {
	return watcher.DEPENDS
}

func (d *Depends) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &Depends{
		Data: make(map[string]map[string]map[string]DependsItem),
	}
	for key, item := range data {
		dd := item.(depends.Depends)
		ret.Data[key] = make(map[string]map[string]DependsItem)
		for nodeName, appList := range dd {
			ret.Data[key][nodeName] = make(map[string]DependsItem)
			for app, item := range appList {
				containers := make([]ContainerInfo, 0, len(item.Pod.Containers))
				for i, _ := range item.Pod.Containers {
					containers = append(containers, ContainerInfo{
						ContainerID: item.Pod.Containers[i].Id,
						NodeIP:      item.Pod.Containers[i].NodeIp,
						IP:          item.Pod.Containers[i].ContainerIp,
						Port:        item.Pod.Containers[i].ContainerPort,
					})
				}
				ret.Data[key][nodeName][app] = DependsItem{
					Annotation: item.Spec.Annotation,
					Containers: containers,
				}
			}
		}
	}
	return ret, !reflect.DeepEqual(ret.Data, d.Data), nil
}

func (d *Depends) Key(r *http.Request) (string, error) {
	if !auth.IsSuper(r.RemoteAddr) {
		return "", fmt.Errorf("authorize failed, super required")
	}
	return api.GetString(r, "target", "*"), nil
}
