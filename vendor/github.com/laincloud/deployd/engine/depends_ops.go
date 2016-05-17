package engine

import (
	"fmt"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/storage"
	"github.com/mijia/adoc"
	"github.com/mijia/go-generics"
	"github.com/mijia/sweb/log"
)

type depOperation interface {
	Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool
}

type depOperStoreRemove struct {
	spec PodSpec
}

func (op depOperStoreRemove) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var err error
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, remove data, err=%v, duration=%s", op.spec, err, time.Now().Sub(start))
	}()
	err = store.Remove(depCtrl.podsStoredKey)
	if err != nil {
		log.Warnf("[Store] Failed to remove depends pods %s, %s", depCtrl.podsStoredKey, err)
	}
	err = store.Remove(depCtrl.specStoredKey)
	if err != nil {
		log.Warnf("[Store] Failed to remove depends spec %s, %s", depCtrl.specStoredKey, err)
	}
	return false
}

type depOperStoreSaveSpec struct {
	spec  PodSpec
	force bool
}

func (op depOperStoreSaveSpec) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var err error
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, save spec into storage, err=%v, duration=%s", op.spec, err, time.Now().Sub(start))
	}()
	err = store.Set(depCtrl.specStoredKey, op.spec, op.force)
	if err != nil {
		log.Warnf("[Store] Failed to save depends pod spec %s, %s", depCtrl.specStoredKey, err)
	}
	return false
}

type depOperStoreSavePods struct {
	spec PodSpec
}

func (op depOperStoreSavePods) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var err error
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, save pods into storage, err=%v, duration=%s", op.spec, err, time.Now().Sub(start))
	}()

	depCtrl.RLock()
	defer depCtrl.RUnlock()

	pods := make(map[string]map[string]SharedPodWithSpec)
	for node, nsPodCtrls := range depCtrl.podCtrls {
		pods[node] = make(map[string]SharedPodWithSpec)
		for namespace, podCtrl := range nsPodCtrls {
			pods[node][namespace] = SharedPodWithSpec{
				RefCount:   podCtrl.refCount,
				VerifyTime: podCtrl.verifyTime,
				Spec:       podCtrl.spec,
				Pod:        podCtrl.pod,
			}
		}
	}
	err = store.Set(depCtrl.podsStoredKey, pods)
	if err != nil {
		log.Warnf("[Store] Failed to save depends pods %s, %s", depCtrl.podsStoredKey, err)
	}
	return false
}

type depOperUpgrade struct {
	newSpec PodSpec
	oldSpec PodSpec
}

func (op depOperUpgrade) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var updateCount int
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, upgrade all instances finished, #updated=%d, duration=%s",
			op.newSpec, updateCount, time.Now().Sub(start))
	}()

	depCtrl.Lock()
	defer depCtrl.Unlock()
	for node, nsPodCtrls := range depCtrl.podCtrls {
		for namespace, podCtrl := range nsPodCtrls {
			upgradeOp := depOperUpgradeInstance{podCtrl, node, namespace, op.newSpec}
			upgradeOp.Do(depCtrl, c, store, ev)
			updateCount++
		}
	}
	return false
}

type depOperUpgradeInstance struct {
	podCtrl   *sharedPodController
	node      string
	namespace string
	newSpec   PodSpec
}

func (op depOperUpgradeInstance) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	podCtrl := op.podCtrl
	prevSpec := podCtrl.spec.Clone()
	prevPod := podCtrl.pod.Clone()

	defer func() {
		log.Infof("DependsCtrl %s upgrade to %s, #namespace=%s, #node=%s",
			prevSpec, op.newSpec, op.namespace, op.node)
	}()

	removeOp := depOperRemoveInstance{podCtrl, prevSpec, prevPod}
	removeOp.Do(depCtrl, c, store, ev)
	time.Sleep(7 * time.Second)

	newSpec := depCtrl.specifyPodSpec(op.newSpec, op.node, op.namespace)
	podCtrl.spec = newSpec
	podCtrl.pod.State = RunStatePending
	deployOp := depOperDeployInstance{podCtrl}
	deployOp.Do(depCtrl, c, store, ev)
	return false
}

type depOperRefresh struct {
	spec PodSpec
}

func (op depOperRefresh) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, refreshed, duration=%s", op.spec, time.Now().Sub(start))
	}()

	depCtrl.Lock()
	defer depCtrl.Unlock()
	for node, nsPodCtrls := range depCtrl.podCtrls {
		for namespace, podCtrl := range nsPodCtrls {
			lastVerify := podCtrl.verifyTime
			if depCtrl.startedAt.Add(kDependsGarbageCollectTimeout).Before(time.Now()) {
				// we have enough time for the podgroups to re verify their depends
				if lastVerify.Add(kDependsGarbageCollectTimeout).Before(time.Now()) {
					log.Warnf("DependsCtrl %s, found pod not verified for a long time, will remove it pod=%s, last verifyTime=%s, refCount=%d",
						op.spec, podCtrl.spec.Name, lastVerify, podCtrl.refCount)
					op := depOperRemoveInstance{podCtrl, podCtrl.spec, podCtrl.pod}
					op.Do(depCtrl, c, store, ev)
					delete(depCtrl.podCtrls[node], namespace)
					continue
				}
			}
			op := depOperRefreshInstance{op.spec, podCtrl, node, namespace}
			op.Do(depCtrl, c, store, ev)
		}
	}
	return false
}

type depOperRemove struct {
	spec PodSpec
}

func (op depOperRemove) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var removeCount int
	start := time.Now()
	defer func() {
		log.Warnf("DependsCtrl %s, remove all instances finished, #removed=%d, duration=%s",
			op.spec, removeCount, time.Now().Sub(start))
	}()

	depCtrl.Lock()
	defer depCtrl.Unlock()
	for _, nsPodCtrls := range depCtrl.podCtrls {
		for _, podCtrl := range nsPodCtrls {
			op := depOperRemoveInstance{podCtrl, podCtrl.spec, podCtrl.pod}
			op.Do(depCtrl, c, store, ev)
			removeCount++
		}
	}
	depCtrl.podCtrls = make(map[string]map[string]*sharedPodController)
	return false
}

type depOperSnapshotEagleView struct {
	spec PodSpec
}

func (op depOperSnapshotEagleView) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var _err error
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, snapshot the eagle view, err=%v, duration=%s", op.spec, _err, time.Now().Sub(start))
	}()

	if pods, err := ev.RefreshPodsByNamespace(c, op.spec.Name); err != nil {
		_err = err
	} else {
		snapshot := make([]RuntimeEaglePod, len(pods))
		copy(snapshot, pods)
		depCtrl.evSnapshot = snapshot
	}
	return false
}

type depOperRemoveInstance struct {
	podCtrl *sharedPodController
	spec    PodSpec
	pod     Pod
}

func (op depOperRemoveInstance) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	podCtrl := op.podCtrl
	podCtrl.Remove(c)
	depCtrl.emitChangeEvent("remove", op.spec, op.pod)
	return false
}

type depOperDeployInstance struct {
	podCtrl *sharedPodController
}

func (op depOperDeployInstance) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	podCtrl := op.podCtrl
	newSpec := podCtrl.spec
	containerIds := make([]string, len(podCtrl.spec.Containers))
	foundDeployed := false
	for _, podContainer := range depCtrl.evSnapshot {
		if podContainer.Name == newSpec.Name && podContainer.Version == newSpec.Version {
			cId := podContainer.Container.Id
			cIndex := podContainer.ContainerIndex
			if cIndex >= 0 && cIndex < len(podCtrl.spec.Containers) {
				containerIds[cIndex] = cId
				foundDeployed = true
			}
		}
	}

	if !foundDeployed {
		podCtrl.pod.State = RunStatePending
		podCtrl.Deploy(c)
		if podCtrl.pod.State == RunStateSuccess {
			depCtrl.emitChangeEvent("add", newSpec, podCtrl.pod.Clone())
		}
	} else {
		log.Warnf("DependsCtrl, we just found pod[%q, version=%d] deployed, refresh the instance and get it back",
			podCtrl.spec.Name, podCtrl.spec.Version)
		podCtrl.pod.Containers = make([]Container, len(containerIds))
		for i, cId := range containerIds {
			podCtrl.pod.Containers[i].Id = cId
		}
		podCtrl.Refresh(c)
		if podCtrl.pod.State == RunStateSuccess {
			depCtrl.emitChangeEvent("verify", newSpec, podCtrl.pod.Clone())
		}
	}
	return false
}

type depOperRefreshInstance struct {
	spec      PodSpec
	podCtrl   *sharedPodController
	node      string
	namespace string
}

func (op depOperRefreshInstance) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var runtime ImRuntime
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, refresh instance on node=%s, namespace=%s, runtime=%+v, duration=%s",
			op.spec, op.node, op.namespace, runtime, time.Now().Sub(start))
	}()

	podCtrl := op.podCtrl
	podCtrl.Refresh(c)
	runtime = podCtrl.pod.ImRuntime

	evIds := make([]string, len(podCtrl.spec.Containers))
	evVersion := -1
	for _, podContainer := range depCtrl.evSnapshot {
		if podContainer.Name == podCtrl.spec.Name {
			evVersion = podContainer.Version
			cId := podContainer.Container.Id
			cIndex := podContainer.ContainerIndex
			if cIndex >= 0 && cIndex < len(evIds) {
				evIds[cIndex] = cId
			}
		}
	}

	if runtime.State == RunStateSuccess {
		if generics.Equal_StringSlice(evIds, podCtrl.pod.ContainerIds()) && podCtrl.spec.Version == evVersion {
			depCtrl.emitChangeEvent("verify", podCtrl.spec, podCtrl.pod.Clone())
			return false
		}
	}

	if runtime.State == RunStateMissing {
		foundRuntime := false
		for i, cId := range evIds {
			if cId != "" {
				podCtrl.pod.Containers[i].Id = cId
				foundRuntime = true
			}
		}
		if foundRuntime {
			// recover from the runtime
			podCtrl.Refresh(c)
			runtime = podCtrl.pod.ImRuntime
			if runtime.State == RunStateSuccess {
				depCtrl.emitChangeEvent("verify", podCtrl.spec, podCtrl.pod.Clone())
			}
		} else {
			log.Warnf("DependsCtrl %s, we found pod missing, just redeploy it", op.spec)
			newSpec := op.spec.Clone()
			newSpec = depCtrl.specifyPodSpec(newSpec, op.node, op.namespace)
			podCtrl.spec = newSpec
			podCtrl.pod.State = RunStatePending
			op := depOperDeployInstance{podCtrl}
			op.Do(depCtrl, c, store, ev)
			runtime = podCtrl.pod.ImRuntime
		}
		return false
	}

	// FIXME - to remove this
	// if we are having pod running without container labels, we should also upgrade them
	// remove this if all our containers run with labels
	if len(podCtrl.pod.Containers) > 0 {
		containerDetail := podCtrl.pod.Containers[0].Runtime
		var label ContainerLabel
		if !label.FromMaps(containerDetail.Config.Labels) || label.Namespace == "" {
			log.Warnf("DependsCtrl %s, we found pod running without labels, just upgrade it", op.spec)
			upgradeOp := depOperUpgradeInstance{podCtrl, op.node, op.namespace, op.spec}
			upgradeOp.Do(depCtrl, c, store, ev)
			runtime = podCtrl.pod.ImRuntime
			return false
		}
	}

	if (evVersion != -1 && op.spec.Version != evVersion) || podCtrl.spec.Version != op.spec.Version {
		log.Warnf("DependsCtrl %s, we found pod running with lower version, just upgrade it", op.spec)
		upgradeOp := depOperUpgradeInstance{podCtrl, op.node, op.namespace, op.spec}
		upgradeOp.Do(depCtrl, c, store, ev)
		runtime = podCtrl.pod.ImRuntime
		return false
	}

	if runtime.State == RunStateExit || runtime.State == RunStateFail {
		log.Warnf("DependsCtrl %s, we found pod down, just restart it", op.spec)
		podCtrl.Start(c)
		runtime = podCtrl.pod.ImRuntime
		if runtime.State == RunStateSuccess {
			depCtrl.emitChangeEvent("verify", podCtrl.spec, podCtrl.pod.Clone())
		}
		return false
	}

	return false
}

type depOperDeployPod struct {
	spec      PodSpec
	namespace string
	nodeName  string
}

func (op depOperDeployPod) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var runtime ImRuntime
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, add new pod, namespace=%s, node=%s, runtime=%+v, duration=%s",
			op.spec, op.namespace, op.nodeName, runtime, time.Now().Sub(start))
	}()
	newSpec := depCtrl.specifyPodSpec(op.spec, op.nodeName, op.namespace)

	depCtrl.Lock()
	defer depCtrl.Unlock()
	var pod Pod
	pod.State = RunStatePending
	podCtrl, isNew := depCtrl.getOrAddPodCtrl(op.nodeName, op.namespace, newSpec, pod)
	podCtrl.refCount++
	podCtrl.verifyTime = time.Now()
	if !isNew {
		runtime = podCtrl.pod.ImRuntime
	} else {
		deployOp := depOperDeployInstance{podCtrl}
		deployOp.Do(depCtrl, c, store, ev)
		runtime = podCtrl.pod.ImRuntime
	}
	return false
}

type depOperRemovePod struct {
	spec      PodSpec
	namespace string
	node      string
}

func (op depOperRemovePod) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	status := "No such pod"
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, remove pod finished, %s, namespace=%s, node=%s, duration=%s",
			op.spec, status, op.namespace, op.node, time.Now().Sub(start))
	}()

	depCtrl.Lock()
	defer depCtrl.Unlock()
	if podCtrl, ok := depCtrl.podCtrls[op.node][op.namespace]; ok {
		podCtrl.refCount--
		podCtrl.verifyTime = time.Now()
		status = fmt.Sprintf("%s", podCtrl)
	}
	return false
}

type depOperVerifyPod struct {
	spec      PodSpec
	namespace string
	node      string
}

func (op depOperVerifyPod) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	status := "No such pod"
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, verify pod finished, %s, namespace=%s, node=%s, duration=%s",
			op.spec, status, op.namespace, op.node, time.Now().Sub(start))
	}()

	depCtrl.Lock()
	defer depCtrl.Unlock()
	if podCtrl, ok := depCtrl.podCtrls[op.node][op.namespace]; ok {
		podCtrl.verifyTime = time.Now()
		status = fmt.Sprintf("%s", podCtrl)
	}
	return false
}

type depOperPurge struct {
	spec PodSpec
}

func (op depOperPurge) Do(depCtrl *dependsController, c cluster.Cluster, store storage.Store, ev *RuntimeEagleView) bool {
	var purgeCount int
	start := time.Now()
	defer func() {
		log.Infof("DependsCtrl %s, purge finished, #purged=%d, duration=%s",
			op.spec, purgeCount, time.Now().Sub(start))
	}()
	for _, podContainer := range depCtrl.evSnapshot {
		cId := podContainer.Container.Id
		log.Warnf("DependsCtrl %s, still find some container alive, try to remove it", op.spec)
		// try to stop container first
		if err := c.StopContainer(cId, op.spec.GetKillTimeout()); err != nil {
			log.Warnf("DependsCtrl %s, fail to stop container %s, %s, will remove it directly", op.spec, cId, err.Error())
		}
		for i := 0; i < 3; i += 1 {
			if err := c.RemoveContainer(cId, true, false); err == nil || adoc.IsNotFound(err) {
				purgeCount++
				break
			} else {
				time.Sleep(10 * time.Second)
				if i == 2 {
					log.Warnf("DependsCtrl %s, still cannot remove the container after max retry, please remove it manually", op.spec)
				}
			}
		}
	}
	depCtrl.Lock()
	defer depCtrl.Unlock()
	depCtrl.removeStatus = 1
	return true
}
