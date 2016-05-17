package watcher

import (
	"strings"
	"sync"
)

// Cacher is a simple cache, used to cache converted data from store
type Cacher struct {
	sync.RWMutex
	data map[string]interface{}
}

// NewCacher create a new cache and initialize it by the given data
func NewCacher(data map[string]interface{}) *Cacher {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &Cacher{
		data: data,
	}
}

// Put set the key in cache, if value is nil, it will delete key in cache
func (c *Cacher) Put(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if value == nil {
		delete(c.data, key)
	} else {
		c.data[key] = value
	}
}

// Reset will the data in cache, all the old data will be deleted, replaced by given new data
func (c *Cacher) Reset(data map[string]interface{}) {
	c.Lock()
	defer c.Unlock()
	c.data = data
}

// Delete delete the key in cache, if recursive is true, all the keys having `key` prefix will be deleted
func (c *Cacher) Delete(key string, recursive bool) {
	c.Lock()
	defer c.Unlock()
	if recursive {
		for k := range c.data {
			if strings.HasPrefix(k, key) {
				delete(c.data, k)
			}
		}
	} else {
		delete(c.data, key)
	}
}

// Get find the values in cache, if key was found, return the map with only one key; or it will return all the KV data which key has `key` prefix
func (c *Cacher) Get(key string) map[string]interface{} {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.data[key]; ok {
		return map[string]interface{}{
			key: v,
		}
	}
	ret := make(map[string]interface{})
	for k, v := range c.data {
		if strings.HasPrefix(k, key) {
			ret[k] = v
		}
	}
	return ret
}

// GetAll return all the data in cache
func (c *Cacher) GetAll() map[string]interface{} {
	c.RLock()
	defer c.RUnlock()
	return c.data
}
