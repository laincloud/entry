package v2

import (
	"encoding/json"
	"fmt"
	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"github.com/laincloud/lainlet/watcher"
	"net/http"
	"reflect"
	"strings"
)

var (
	secretKeys = []string{
		"*",
		"swarm_manager_ip",
		"super_apps",
		"dnsmasq_servers",
		"calico_default_rule",
		"calico_network",
		"dnsmasq_addresses",
		"ssl",
		"vips",
		"tinydns_fqdns",
		"bootstrap_node_ip",
		"dns_port",
		"vip",
		"etcd_cluster_token",
		"system_volumes",
		"rsyncd_secrets",
		"dns_ip",
		"node_network",
	}
)

func isSecret(key string) bool {
	for _, sk := range secretKeys {
		if strings.HasPrefix(key, sk) {
			return true
		}
	}
	return false
}

// Config API
type GeneralConfig struct {
	Data map[string]string // data type return by configwatcher
}

func (gc *GeneralConfig) Decode(r []byte) error {
	return json.Unmarshal(r, &gc.Data)
}

func (gc *GeneralConfig) Encode() ([]byte, error) {
	return json.Marshal(gc.Data)
}

func (gc *GeneralConfig) URI() string {
	return "/configwatcher"
}

func (gc *GeneralConfig) WatcherName() string {
	return watcher.CONFIG
}

func (gc *GeneralConfig) Make(conf map[string]interface{}) (api.API, bool, error) {
	ret := &GeneralConfig{
		Data: make(map[string]string),
	}
	for k, v := range conf {
		ret.Data[k], _ = v.(string)
	}

	return ret, !reflect.DeepEqual(gc.Data, ret.Data), nil
}

func (gc *GeneralConfig) Key(r *http.Request) (string, error) {
	target := api.GetString(r, "target", "*")
	if isSecret(target) && !auth.IsSuper(r.RemoteAddr) {
		return "", fmt.Errorf("authorize failed, super required")
	}
	return target, nil
}
