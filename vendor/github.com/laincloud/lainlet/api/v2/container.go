package v2

import (
	"encoding/json"
	"fmt"
	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/container"
	"net/http"
)

// node watcher api, /lain/nodes/nodes
type GeneralContainers struct {
	Data map[string]container.Info // data type return by configwatcher
}

func (gc *GeneralContainers) Decode(r []byte) error {
	return json.Unmarshal(r, &gc.Data)
}

func (gc *GeneralContainers) Encode() ([]byte, error) {
	return json.Marshal(gc.Data)
}

func (gc *GeneralContainers) URI() string {
	return "/containers"
}

func (gc *GeneralContainers) WatcherName() string {
	return watcher.CONTAINER
}

func (gc *GeneralContainers) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &GeneralContainers{
		Data: make(map[string]container.Info),
	}
	for k, v := range data {
		ret.Data[k] = v.(container.Info)
	}
	return ret, true, nil
}

func (gc *GeneralContainers) Key(r *http.Request) (string, error) {
	if !auth.IsSuper(r.RemoteAddr) {
		return "", fmt.Errorf("authorize failed, super required")
	}
	target := api.GetString(r, "nodename", "*")
	return target, nil
}
