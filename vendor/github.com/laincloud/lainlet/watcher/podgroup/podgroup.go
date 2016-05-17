package podgroup

import (
	"encoding/json"
	"fmt"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/deployd/engine"
	"github.com/laincloud/lainlet/store"
	"github.com/laincloud/lainlet/watcher"
	"path"
)

const (
	// KEY represents the key path in store
	KEY = "/lain/deployd/pod_groups"
)

// PodGroup represents the data type stored in backend for each pod. watcher will return data whose type is map[string]PodGroup.
type PodGroup engine.PodGroupWithSpec

// New create a new watcher which used to watch podgroup data
func New(s store.Store, ctx context.Context) (*watcher.Watcher, error) {
	return watcher.New(s, ctx, KEY, convert, invertKey)
}

func invertKey(key string) string {
	return path.Join(KEY, key)
}

func convert(pairs []*store.KVPair) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	for _, kv := range pairs {
		var pg PodGroup
		if err := json.Unmarshal(kv.Value, &pg); err != nil {
			log.Errorf("Fail to unmarshal the podgroup data: %s", string(kv.Value))
			log.Errorf("JSON unmarshal error: %s", err.Error())
			return nil, fmt.Errorf("a KVPair unmarshal failed")
		}
		ret[kv.Key[len(KEY)+1:]] = pg
	}
	return ret, nil
}
