package depends

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
	KEY = "/lain/deployd/depends/pods"
)

// Depends represents the data type returned by this watcher. in fact, it's a type to represents map[nodename]map[appname]SharedPodWithSpec
type Depends map[string]map[string]engine.SharedPodWithSpec

// New create a new watcher which used to watch depends data
func New(s store.Store, ctx context.Context) (*watcher.Watcher, error) {
	return watcher.New(s, ctx, KEY, convert, invertKey)
}

func invertKey(key string) string {
	return path.Join(KEY, key)
}

func convert(pairs []*store.KVPair) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	for _, kv := range pairs {
		var dp Depends
		if err := json.Unmarshal(kv.Value, &dp); err != nil {
			log.Errorf("Fail to unmarshal the dependency data: %s", string(kv.Value))
			log.Errorf("JSON unmarshal error: %s", err.Error())
			return nil, fmt.Errorf("a KVPair unmarshal failed")
		}
		ret[kv.Key[len(KEY)+1:]] = dp
	}
	return ret, nil
}
