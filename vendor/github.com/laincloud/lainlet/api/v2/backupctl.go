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

type ContainerForBackupctl struct {
	Id       string
	Ip       string
	NodeIp   string
	NodeName string
}

type PodInfoForBackupctl struct {
	Annotation string
	Containers []ContainerForBackupctl
	InstanceNo int
}

type CoreInfoForBackupctl struct {
	Data map[string][]PodInfoForBackupctl
}

func (ci *CoreInfoForBackupctl) Decode(r []byte) error {
	return json.Unmarshal(r, &ci.Data)
}

func (ci *CoreInfoForBackupctl) Encode() ([]byte, error) {
	return json.Marshal(ci.Data)
}

func (ci *CoreInfoForBackupctl) URI() string {
	return "/backupspec"
}

func (ci *CoreInfoForBackupctl) WatcherName() string {
	return watcher.PODGROUP
}

func (ci *CoreInfoForBackupctl) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &CoreInfoForBackupctl{
		Data: make(map[string][]PodInfoForBackupctl),
	}
	for _, item := range data {
		pg := item.(podgroup.PodGroup)
		infos := make([]PodInfoForBackupctl, len(pg.Pods))
		for i, pod := range pg.Pods {
			infos[i] = PodInfoForBackupctl{
				Annotation: pg.Spec.Pod.Annotation,
				InstanceNo: pod.InstanceNo,
				Containers: make([]ContainerForBackupctl, len(pod.Containers)),
			}
			for j, container := range pod.Containers {
				infos[i].Containers[j] = ContainerForBackupctl{
					Id:       container.Id,
					Ip:       container.ContainerIp,
					NodeIp:   container.NodeIp,
					NodeName: container.NodeName,
				}
			}
		}
		ret.Data[pg.Spec.Name] = infos
	}
	return ret, !reflect.DeepEqual(ci.Data, ret.Data), nil
}

func (ci *CoreInfoForBackupctl) Key(r *http.Request) (string, error) {
	if !auth.IsSuper(r.RemoteAddr) {
		return "", fmt.Errorf("authorize failed, super required")
	}
	appName := api.GetString(r, "appname", "*")
	return appName, nil
}
