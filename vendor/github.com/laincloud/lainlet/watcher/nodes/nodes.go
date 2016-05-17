package nodes

import (
	"encoding/json"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/lainlet/store"
	"github.com/laincloud/lainlet/watcher"
	"path"
)

const (
	// KEY represents the key path in store
	KEY = "/lain/nodes/nodes"
)

// NodeInfo represents the data type returned by this watcher. it's represents map[nodeip/nodename]string(or map)
type NodeInfo map[string]interface{}

// New create a new watcher which used to watch node info
func New(s store.Store, ctx context.Context) (*watcher.Watcher, error) {
	return watcher.New(s, ctx, KEY, convert, invertKey)
}

func invertKey(key string) string {
	return path.Join(KEY, key)
}

func convert(pairs []*store.KVPair) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	for _, kv := range pairs {
		var tmp NodeInfo
		if err := json.Unmarshal(kv.Value, &tmp); err != nil {
			log.Errorf("Fail to unmarshal nodes data %v", string(kv.Value))
			continue
		}
		ret[kv.Key[len(KEY)+1:]] = tmp
	}
	return ret, nil
}
