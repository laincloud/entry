package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/mijia/adoc"
	"github.com/mijia/sweb/log"
)

type RuntimeEaglePod struct {
	ContainerLabel
	Container   adoc.Container
	ParseSource string
}

func (pod RuntimeEaglePod) String() string {
	return fmt.Sprintf("<RuntimeEaglePod name=%s, instance=%d, version=%d, drift=%d, cIndex=%d, cId=%s, source=%s>",
		pod.Name, pod.InstanceNo, pod.Version, pod.DriftCount, pod.ContainerIndex,
		pod.Container.Id[:12], pod.ParseSource)
}

type RuntimeEagleView struct {
	sync.RWMutex
	podGroups map[string][]RuntimeEaglePod
}

func (ev *RuntimeEagleView) GetRuntimeEaglePods(name string) ([]RuntimeEaglePod, bool) {
	ev.RLock()
	defer ev.RUnlock()
	pods, ok := ev.podGroups[name]
	return pods, ok
}

func (ev *RuntimeEagleView) Refresh(c cluster.Cluster) error {
	totalContainers, totalPodGroups := 0, 0
	start := time.Now()
	defer func() {
		log.Infof("<RuntimeEagleView> refreshed, podContainers=%d, podGroups=%d, duration=%s",
			totalContainers, totalPodGroups, time.Now().Sub(start))
	}()

	labelFilter := []string{"com.docker.swarm.id"}
	podGroups := make(map[string][]RuntimeEaglePod)
	err := ev.refreshCallback(c, labelFilter, func(pod RuntimeEaglePod) {
		name := pod.Name
		podGroups[name] = append(podGroups[name], pod)
		totalContainers += 1
	})
	if err == nil {
		ev.Lock()
		defer ev.Unlock()
		ev.podGroups = podGroups
		totalPodGroups = len(podGroups)
	}
	return err
}

func (ev *RuntimeEagleView) RefreshPodGroup(c cluster.Cluster, pgName string) ([]RuntimeEaglePod, error) {
	totalContainers := 0
	start := time.Now()
	defer func() {
		log.Infof("<RuntimeEagleView> pod group %s refreshed, podContainers=%d, duration=%s",
			pgName, totalContainers, time.Now().Sub(start))
	}()

	filters := []string{
		fmt.Sprintf("%s.pg_name=%s", kLainLabelPrefix, pgName),
	}
	pods, err := ev.refreshByFilters(c, filters)
	totalContainers = len(pods)
	return pods, err
}

func (ev *RuntimeEagleView) RefreshPodsByNamespace(c cluster.Cluster, namespace string) ([]RuntimeEaglePod, error) {
	totalContainers := 0
	start := time.Now()
	defer func() {
		log.Infof("<RuntimeEagleView> pods by namespace %s refreshed, #containers=%d, duration=%s",
			namespace, totalContainers, time.Now().Sub(start))
	}()

	filters := []string{
		fmt.Sprintf("%s.pg_namespace=%s", kLainLabelPrefix, namespace),
	}
	pods, err := ev.refreshByFilters(c, filters)
	totalContainers = len(pods)
	return pods, err
}

func (ev *RuntimeEagleView) refreshByFilters(c cluster.Cluster, labelFilters []string) ([]RuntimeEaglePod, error) {
	labelFilters = append(labelFilters, "com.docker.swarm.id")
	pods := make([]RuntimeEaglePod, 0, 10)
	err := ev.refreshCallback(c, labelFilters, func(pod RuntimeEaglePod) {
		pods = append(pods, pod)
	})
	return pods, err
}

func (ev *RuntimeEagleView) refreshCallback(c cluster.Cluster, labelFilter []string, callback func(RuntimeEaglePod)) error {
	filters := map[string][]string{
		"label": labelFilter,
	}
	filterJson, err := json.Marshal(filters)
	if err != nil {
		log.Warnf("<RuntimeEagleView> Failed to encode the filter json, %s", err)
		return err
	}
	if containers, err := c.ListContainers(true, false, string(filterJson)); err != nil {
		log.Warnf("<RuntimeEagleView> Failed to list all containers from swarm, %s", err)
		return err
	} else {
		for _, container := range containers {
			var pod *RuntimeEaglePod
			if _pod, ok := ev.extractFromLabel(container); ok {
				pod = &_pod
			} else if _pod, ok := ev.extractFromName(container); ok {
				pod = &_pod
			}
			if pod != nil {
				log.Debugf("Found runtime eagle pod container, %s", pod)
				callback(*pod)
			}
		}
		return nil
	}
}

func (ev *RuntimeEagleView) Activate(c cluster.Cluster) {
	// FIXME do nothing for now, don't know if we need to refresh this yet
	// or should we monitor the cluster event
	// go ev.startEventMonitor(c)
}

func (ev *RuntimeEagleView) startEventMonitor(c cluster.Cluster) {
	// FIXME nothing here yet
}

func (ev *RuntimeEagleView) extractFromLabel(container adoc.Container) (RuntimeEaglePod, bool) {
	var pod RuntimeEaglePod
	if ok := pod.ContainerLabel.FromMaps(container.Labels); !ok {
		return pod, false
	}
	pod.Container = container
	pod.ParseSource = "label"
	return pod, true
}

var (
	// name is like "node1/deploy.web.web.v0-i1-d0" or "/deploy.web.web.v0-i1-d0-c0"
	// also we may have "node1/deploy.web.web.v0-i1-d0-lain_did_it_ddddddddd" to rename the conflict one which we should not create those anymore
	lainContainerNamePattern = regexp.MustCompile("\\.v([0-9]+)-i([0-9]+)-d([0-9]+)(-c([0-9]+))*$")
)

func (ev *RuntimeEagleView) extractFromName(container adoc.Container) (RuntimeEaglePod, bool) {
	var pod RuntimeEaglePod
	if len(container.Names) == 0 {
		return pod, false
	}
	parts := strings.Split(container.Names[0], "/")
	name := parts[len(parts)-1]
	matches := lainContainerNamePattern.FindStringSubmatch(name)
	if len(matches) != 6 {
		return pod, false
	}
	var err error
	pod.Name = strings.TrimSuffix(name, matches[0])
	hasError := pod.Name == ""
	pod.Version, err = strconv.Atoi(matches[1])
	hasError = hasError || err != nil
	pod.InstanceNo, err = strconv.Atoi(matches[2])
	hasError = hasError || err != nil
	pod.DriftCount, err = strconv.Atoi(matches[3])
	hasError = hasError || err != nil
	if matches[5] != "" {
		pod.ContainerIndex, err = strconv.Atoi(matches[5])
		hasError = hasError || err != nil
	}
	pod.Container = container
	pod.ParseSource = "name"
	return pod, !hasError
}

func NewRuntimeEagleView() *RuntimeEagleView {
	ev := &RuntimeEagleView{
		podGroups: make(map[string][]RuntimeEaglePod),
	}
	return ev
}
