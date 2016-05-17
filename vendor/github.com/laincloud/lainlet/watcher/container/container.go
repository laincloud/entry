package container

import (
	"encoding/json"
	"fmt"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/deployd/engine"
	"github.com/laincloud/lainlet/store"
	"github.com/laincloud/lainlet/watcher"
)

const (
	// KEY represents the key path in store
	KEY = "/lain/deployd/pod_groups"
)

var (
	invertsTable map[string]string
)

func init() {
	invertsTable = make(map[string]string)
}

// PodGroup is actually from the deployd engine, it actually engine.PodGroupWithSpec
type PodGroup engine.PodGroupWithSpec

// Info represents the container info, the data type returned by this container watcher
type Info struct {
	AppName    string `json:"app"`
	ProcName   string `json:"proc"`
	NodeName   string `json:"nodename"`
	NodeIP     string `json:"nodeip"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	InstanceNo int    `json:"instanceNo"`
}

// New create a new watcher which used to watch container info
func New(s store.Store, ctx context.Context) (*watcher.Watcher, error) {
	return watcher.New(s, ctx, KEY, convert, invertKey)
}

func invertKey(key string) string {
	s, ok := invertsTable[key]
	if ok {
		return s
	}
	return ""
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

		/*
		 * we should delete old data in cache and invertTable.
		 * the etcd is set by deployd when app upgraded, if we do not delete here,
		 * the cache will having some containers which have been removed after upgrade
		 */
		deletedKeys := make([]string, 0, 2)
		for k, v := range invertsTable { // find the key should be deleted in invertsTable by the etcd key
			if v == kv.Key {
				deletedKeys = append(deletedKeys, k)
			}
		}
		for _, k := range deletedKeys {
			ret[k] = nil            // set to nil, let watcher delete it in cache
			delete(invertsTable, k) // delete it in invert table
		}

		for _, pod := range pg.Pods {
			for _, container := range pod.Containers {
				k1 := fmt.Sprintf("%s/%s", container.NodeName, container.Id)
				k2 := fmt.Sprintf("%s/%s", container.NodeIp, container.Id)
				ci := Info{
					AppName:    pg.Spec.Namespace,
					ProcName:   pg.Spec.Name,
					NodeName:   container.NodeName,
					NodeIP:     container.NodeIp,
					IP:         container.ContainerIp,
					Port:       container.ContainerPort,
					InstanceNo: pod.InstanceNo,
				}

				invertsTable[k1] = kv.Key // lock in case of concurrent ?
				invertsTable[k2] = kv.Key // lock in case of concurrent ?
				ret[k1], ret[k2] = ci, ci
			}
		}

	}
	return ret, nil
}
