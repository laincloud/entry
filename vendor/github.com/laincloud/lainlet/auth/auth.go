package auth

import (
	"encoding/json"
	"fmt"
	"github.com/laincloud/deployd/engine"
	"github.com/laincloud/lainlet/store"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/depends"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"path"
	"strings"
)

const (
	dependsKey        = "/lain/deployd/depends/pods"
	podgroupKey       = "/lain/deployd/pod_groups"
	superAppsStoreKey = "/lain/config/super_apps"
)

var (
	active              = true
	localIP             = "127.0.0.1"
	authTable           map[string]containerInfo
	dependsInvertTable  map[string]string
	podgroupInvertTable map[string]string
	dependsWatcher      *watcher.Watcher
	podgroupWatcher     *watcher.Watcher
	superAppsWatcher    *watcher.Watcher
)

// Init function initialize the data needed by auth.
func Init(s store.Store, ctx context.Context, ip string, act bool) error {
	active = act // set active, auth always return success when unactive
	localIP = ip
	authTable = make(map[string]containerInfo)
	dependsInvertTable = make(map[string]string)
	podgroupInvertTable = make(map[string]string)
	var err error
	dependsWatcher, err = watcher.New(s, ctx, dependsKey, dependsConvert, dependsInvertKey)
	if err != nil {
		return err
	}
	podgroupWatcher, err = watcher.New(s, ctx, podgroupKey, podgroupConvert, podgroupInvertKey)
	if err != nil {
		return err
	}
	superAppsWatcher, err = watcher.New(s, ctx, superAppsStoreKey, superAppsConvert, superAppsInvertKey)
	if err != nil {
		return err
	}
	return nil
}

type containerInfo struct {
	AppName     string
	Proc        string
	ContainerID string
}

func isSuperApp(appname string) bool {
	if !active {
		log.Debugf("auth is not active, return true")
		return true
	}
	if data, _ := superAppsWatcher.Get(appname); data != nil && len(data) == 1 {
		return true
	}
	log.Warnf("app %s is not a super apps, not found it in super configuration", appname)
	return false
}

// IsSuper check if the ip belong to the super apps
func IsSuper(remoteIP string) bool {
	if !active {
		log.Debugf("auth is not active, return true")
		return true
	}
	if index := strings.LastIndexByte(remoteIP, ':'); index >= 0 {
		remoteIP = remoteIP[:index]
	}
	log.Debugf("remote ip is %s", remoteIP)
	if remoteIP == "127.0.0.1" || remoteIP == localIP || remoteIP == "[::1]" {
		return true
	}
	info, err := podgroupWatcher.Get(remoteIP)
	if err != nil || info == nil {
		log.Warnf("can not get container info by ip %s, %s", remoteIP, err.Error())
		return false
	}
	if ci, ok := info[remoteIP]; ok {
		if isSuperApp(ci.(containerInfo).AppName) {
			return true
		}
	}
	return false
}

// Pass check if the given ip having limits to visit data for given app
func Pass(remoteIP string, appname string) bool {
	if !active {
		log.Debugf("auth is not active, return true")
		return true
	}
	log.Debugf("verify if %s having limit visiting %s", remoteIP, appname)
	if index := strings.LastIndexByte(remoteIP, ':'); index >= 0 {
		remoteIP = remoteIP[:index]
	}
	log.Debugf("remote ip is %s", remoteIP)
	switch remoteIP {
	case "127.0.0.1", localIP, "[::1]": // ipv6 is [::1]
		return true
	default:
		info, err := podgroupWatcher.Get(remoteIP)
		if err != nil || info == nil {
			log.Debugf("verify failed, %s", err.Error())
			return false
		}

		remoteAppName, err := AppName(remoteIP)
		if err != nil {
			log.Errorf("Fail to get appname by ip %s, %s", remoteIP, err.Error())
			return false
		}

		log.Debugf("check if app %s has rights visiting %s", remoteAppName, appname)
		// visit itself?
		if remoteAppName == appname {
			return true
		}
		// request from a super app?
		if isSuperApp(remoteAppName) {
			return true
		}
		// visit it's dependency service?
		apps, err := dependsWatcher.Get(remoteIP)
		if err != nil {
			log.Debugf("verify failed, %s", err.Error())
			return false
		}
		if slice, ok := apps[remoteIP]; ok {
			for _, item := range slice.([]string) {
				if item == appname {
					return true
				}
			}
		}
	}
	log.Warnf("verify failed, %s has no permission to visit %s", remoteIP, appname)
	return false
}

// AppName return the app name which ip is given ip. the return error is nil only when appname is found
func AppName(remoteIP string) (string, error) {
	if index := strings.LastIndexByte(remoteIP, ':'); index >= 0 {
		remoteIP = remoteIP[:index]
	}
	podgroupData, err := podgroupWatcher.Get(remoteIP)
	if err != nil {
		return "", err
	}
	if appinfo, ok := podgroupData[remoteIP]; ok {
		return appinfo.(containerInfo).AppName, nil
	}
	dependsData, err := dependsWatcher.Get(remoteIP)
	if err != nil {
		return "", err
	}
	if serviceNames, ok := dependsData[remoteIP]; ok && len(serviceNames.([]string)) > 0 {
		return serviceNames.([]string)[0], nil
	}
	return "", fmt.Errorf("unkown address %s", remoteIP)
}

func podgroupInvertKey(key string) string {
	k, ok := podgroupInvertTable[key]
	if ok {
		return k
	}
	return ""
}

func podgroupConvert(pairs []*store.KVPair) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	for _, kv := range pairs {
		var pg engine.PodGroupWithSpec
		if err := json.Unmarshal(kv.Value, &pg); err != nil {
			log.Errorf("Fail to unmarshal the dependency data: %s", string(kv.Value))
			log.Errorf("JSON unmarshal error: %s", err.Error())
			return nil, fmt.Errorf("a KVPair unmarshal failed")
		}
		for _, pod := range pg.Pods {
			for _, container := range pod.Containers {
				ret[container.ContainerIp] = containerInfo{
					AppName:     pg.Spec.Namespace,
					Proc:        pg.Spec.Name,
					ContainerID: container.Id,
				}
				podgroupInvertTable[container.ContainerIp] = kv.Key
			}
		}
	}
	return ret, nil
}

func dependsInvertKey(key string) string {
	k, ok := dependsInvertTable[key]
	if ok {
		return k
	}
	return ""
}

func dependsConvert(pairs []*store.KVPair) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	for _, kv := range pairs {
		var dp depends.Depends
		if err := json.Unmarshal(kv.Value, &dp); err != nil {
			log.Errorf("Fail to unmarshal the dependency data: %s", string(kv.Value))
			log.Errorf("JSON unmarshal error: %s", err.Error())
			return nil, fmt.Errorf("a KVPair unmarshal failed")
		}
		fields := strings.Split(kv.Key[len(dependsKey)+1:], ".")
		serviceName := fields[0]
		if len(fields) > 3 {
			serviceName = strings.Join(fields[:len(fields)-2], ".")
		}
		for _, nodeData := range dp {
			for _, appData := range nodeData {
				for _, container := range appData.Pod.Containers {
					if _, ok := ret[container.ContainerIp]; ok {
						ret[container.ContainerIp] = append(ret[container.ContainerIp].([]string), serviceName)
					} else {
						ret[container.ContainerIp] = []string{serviceName}
					}
					dependsInvertTable[container.ContainerIp] = kv.Key
				}
			}
		}
	}
	return ret, nil
}

func superAppsInvertKey(key string) string {
	return path.Join(superAppsStoreKey, key)
}

func superAppsConvert(pairs []*store.KVPair) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	for _, kv := range pairs {
		ret[kv.Key[len(superAppsStoreKey)+1:]] = string(kv.Value)
	}
	return ret, nil
}
