package v2

import (
	"encoding/json"
	"fmt"
	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/nodes"
	"net/http"
	"reflect"
)

// node watcher api, /lain/nodes/nodes
type GeneralNodes struct {
	Data map[string]nodes.NodeInfo // data type return by configwatcher
}

func (gn *GeneralNodes) Decode(r []byte) error {
	return json.Unmarshal(r, &gn.Data)
}

func (gn *GeneralNodes) Encode() ([]byte, error) {
	return json.Marshal(gn.Data)
}

func (gn *GeneralNodes) URI() string {
	return "/nodes"
}

func (gn *GeneralNodes) WatcherName() string {
	return watcher.NODES
}

func (gn *GeneralNodes) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &GeneralNodes{
		Data: make(map[string]nodes.NodeInfo),
	}
	for k, v := range data {
		ret.Data[k] = v.(nodes.NodeInfo)
	}
	return ret, !reflect.DeepEqual(gn.Data, ret.Data), nil
}

func (gn *GeneralNodes) Key(r *http.Request) (string, error) {
	if !auth.IsSuper(r.RemoteAddr) {
		return "", fmt.Errorf("authorize failed, super required")
	}
	target := api.GetString(r, "name", "*")
	return target, nil
}
