package v2

import (
	"encoding/json"
	"fmt"
	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/container"
	"net/http"
	"reflect"
)

// Localspec API, it do not support watch request
type LocalSpec struct {
	Data    []string
	LocalIP string
	ip      string
}

func (ls *LocalSpec) Decode(r []byte) error {
	return json.Unmarshal(r, &ls.Data)
}

func (ls *LocalSpec) Encode() ([]byte, error) {
	return json.Marshal(ls.Data)
}

func (ls *LocalSpec) URI() string {
	return "/localspecquery"
}

func (ls *LocalSpec) WatcherName() string {
	return watcher.CONTAINER
}

func (ls *LocalSpec) Make(data map[string]interface{}) (api.API, bool, error) {
	ret := &LocalSpec{
		Data: make([]string, 0),
	}
	// merge the repeat item, so using map
	set := make(map[string]bool)
	for _, item := range data {
		ci := item.(container.Info)
		set[fmt.Sprintf("%s/%s", ci.AppName, ci.ProcName)] = true
	}
	for k, _ := range set {
		ret.Data = append(ret.Data, k)
	}
	return ret, !reflect.DeepEqual(ret.Data, ls.Data), nil
}

func (ls *LocalSpec) Key(r *http.Request) (string, error) {
	if !auth.IsSuper(r.RemoteAddr) {
		return "", fmt.Errorf("authorize failed, super required")
	}
	return api.GetString(r, "nodeip", ls.LocalIP), nil
}

// to realize BanWatcher interface, abandon watch action
func (ls *LocalSpec) BanWatch() {}
