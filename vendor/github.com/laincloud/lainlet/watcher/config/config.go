package config

import (
	"golang.org/x/net/context"
	"github.com/laincloud/lainlet/store"
	"github.com/laincloud/lainlet/watcher"
	"path"
)

const (
	// KEY represents the key path in store
	KEY = "/lain/config"
)

// New create a new watcher which watch KEY in backend store
func New(s store.Store, ctx context.Context) (*watcher.Watcher, error) {
	return watcher.New(s, ctx, KEY, convert, invertKey)
}

func invertKey(key string) string {
	return path.Join(KEY, key)
}

func convert(pairs []*store.KVPair) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	for _, kv := range pairs {
		ret[kv.Key[len(KEY)+1:]] = string(kv.Value)
	}
	return ret, nil
}
