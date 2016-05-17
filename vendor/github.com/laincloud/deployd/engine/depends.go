package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/storage"
)

const (
	// set GarbageCollectTimeout long enough.
	// sometimes, podgroup refresh goroutine may having some problem(such as swarm exception),
	// and do not verify it's portal for a long time. 5 hours is enough
	kDependsGarbageCollectTimeout = 5 * time.Hour
)

type NamespacePodsWithSpec struct {
	Spec PodSpec
	Pods map[string][]Pod
}

type SharedPodWithSpec struct {
	RefCount   int
	VerifyTime time.Time
	Spec       PodSpec
	Pod        Pod
}

type sharedPodController struct {
	podController
	refCount   int
	verifyTime time.Time
}

func (podCtrl *sharedPodController) String() string {
	return fmt.Sprintf("SharedPodCtrl %s, refCount=%d, verify=%s",
		podCtrl.spec, podCtrl.refCount, podCtrl.verifyTime)
}

type dependsController struct {
	sync.RWMutex
	spec         PodSpec
	podCtrls     map[string]map[string]*sharedPodController // [node][namespace]podCtrl
	removeStatus int

	Publisher
	evSnapshot    []RuntimeEaglePod
	opsChan       chan depOperation
	startedAt     time.Time
	specStoredKey string
	podsStoredKey string
}

func (depCtrl *dependsController) String() string {
	return fmt.Sprintf("DependsCtrl %s", depCtrl.spec)
}

func (depCtrl *dependsController) RemoveStatus() int {
	depCtrl.RLock()
	defer depCtrl.RUnlock()
	return depCtrl.removeStatus
}

func (depCtrl *dependsController) Inspect() NamespacePodsWithSpec {
	depCtrl.RLock()
	defer depCtrl.RUnlock()

	podsWithSpec := NamespacePodsWithSpec{
		Spec: depCtrl.spec,
		Pods: make(map[string][]Pod),
	}
	for _, nsPodCtrls := range depCtrl.podCtrls {
		for namespace, podCtrl := range nsPodCtrls {
			pods, ok := podsWithSpec.Pods[namespace]
			if !ok {
				pods = make([]Pod, 0, 10)
			}
			pods = append(pods, podCtrl.pod)
			podsWithSpec.Pods[namespace] = pods
		}
	}
	return podsWithSpec
}

func (depCtrl *dependsController) Refresh() {
	depCtrl.RLock()
	spec := depCtrl.spec.Clone()
	depCtrl.RUnlock()
	depCtrl.opsChan <- depOperSnapshotEagleView{spec}
	depCtrl.opsChan <- depOperRefresh{spec}
	depCtrl.opsChan <- depOperStoreSavePods{spec}
}

func (depCtrl *dependsController) AddSpec() {
	depCtrl.RLock()
	spec := depCtrl.spec.Clone()
	depCtrl.RUnlock()
	depCtrl.opsChan <- depOperStoreSaveSpec{spec, true}
}

func (depCtrl *dependsController) UpdateSpec(newSpec PodSpec) {
	toUpdate := false
	var (
		oldSpec   PodSpec
		mergeSpec PodSpec
	)
	depCtrl.Lock()
	if !depCtrl.spec.Equals(newSpec) {
		toUpdate = true
		oldSpec = depCtrl.spec.Clone()
		depCtrl.spec = depCtrl.spec.Merge(newSpec)
		mergeSpec = depCtrl.spec.Clone()
	}
	depCtrl.Unlock()

	if !toUpdate {
		return
	}
	depCtrl.opsChan <- depOperSnapshotEagleView{mergeSpec}
	depCtrl.opsChan <- depOperStoreSaveSpec{mergeSpec, true}
	depCtrl.opsChan <- depOperUpgrade{mergeSpec, oldSpec}
	depCtrl.opsChan <- depOperStoreSavePods{mergeSpec}
}

func (depCtrl *dependsController) RemoveSpec(force bool) {
	depCtrl.Lock()
	depCtrl.removeStatus = 0
	toRemove := force
	if !force {
		for _, nsPodCtrls := range depCtrl.podCtrls {
			for _, podCtrl := range nsPodCtrls {
				if podCtrl.refCount > 0 {
					toRemove = false
					break
				}
			}
			if !toRemove {
				break
			}
		}
	}
	if !toRemove {
		depCtrl.removeStatus = 2
	}
	spec := depCtrl.spec.Clone()
	depCtrl.Unlock()

	if !toRemove {
		return
	}
	depCtrl.opsChan <- depOperStoreRemove{spec}
	depCtrl.opsChan <- depOperRemove{spec}
	depCtrl.opsChan <- depOperSnapshotEagleView{spec}
	depCtrl.opsChan <- depOperPurge{spec}
}

func (depCtrl *dependsController) AddPod(namespace, nodeName string) {
	depCtrl.RLock()
	spec := depCtrl.spec.Clone()
	depCtrl.RUnlock()

	depCtrl.opsChan <- depOperSnapshotEagleView{spec}
	depCtrl.opsChan <- depOperDeployPod{spec, namespace, nodeName}
	depCtrl.opsChan <- depOperStoreSavePods{spec}
}

func (depCtrl *dependsController) RemovePod(namespace, nodeName string) {
	depCtrl.RLock()
	spec := depCtrl.spec.Clone()
	depCtrl.RUnlock()
	depCtrl.opsChan <- depOperRemovePod{spec, namespace, nodeName}
	depCtrl.opsChan <- depOperStoreSavePods{spec}
}

func (depCtrl *dependsController) VerifyPod(namespace, nodeName string) {
	depCtrl.RLock()
	spec := depCtrl.spec.Clone()
	depCtrl.RUnlock()
	depCtrl.opsChan <- depOperVerifyPod{spec, namespace, nodeName}
	depCtrl.opsChan <- depOperStoreSavePods{spec}
}

func (depCtrl *dependsController) Activate(c cluster.Cluster, store storage.Store, eagle *RuntimeEagleView, stop chan struct{}) {
	go func() {
		for {
			select {
			case op := <-depCtrl.opsChan:
				if op.Do(depCtrl, c, store, eagle) {
					return
				}
			case <-stop:
				if len(depCtrl.opsChan) == 0 {
					return
				}
			}
		}
	}()
}

func (depCtrl *dependsController) getOrAddPodCtrl(nodeName string, namespace string, spec PodSpec, pod Pod) (*sharedPodController, bool) {
	if _, ok := depCtrl.podCtrls[nodeName]; !ok {
		depCtrl.podCtrls[nodeName] = make(map[string]*sharedPodController)
	}
	if podCtrl, ok := depCtrl.podCtrls[nodeName][namespace]; ok {
		return podCtrl, false
	} else {
		podCtrl = &sharedPodController{
			podController: podController{
				spec: spec,
				pod:  pod,
			},
		}
		podCtrl.podController.spec.PrevState = NewPodPrevState(1) // new pod controller, create new empty prev state
		depCtrl.podCtrls[nodeName][namespace] = podCtrl
		return podCtrl, true
	}
}

func (depCtrl *dependsController) specifyPodSpec(spec PodSpec, nodeName, namespace string) PodSpec {
	newContainers := make([]ContainerSpec, 0, len(spec.Containers))
	for _, container := range spec.Containers {
		newEnv := make([]string, 0, len(container.Env))
		for _, env := range container.Env {
			newEnv = append(newEnv, env)
		}
		container.Env = newEnv
		newContainers = append(newContainers, container)
	}
	spec.Containers = newContainers
	spec.Namespace = spec.Name
	spec.PrevState = NewPodPrevState(1)
	spec.Name = fmt.Sprintf("%s-%s-%s", spec.Name, nodeName, namespace)
	newFilters := make([]string, 0, len(spec.Filters))
	for _, filter := range spec.Filters {
		if strings.HasPrefix(filter, "constraint:node==") {
			continue
		}
		newFilters = append(newFilters, filter)
	}
	newFilters = append(newFilters, fmt.Sprintf("constraint:node==%s", nodeName))
	spec.Filters = newFilters
	return spec
}

func (depCtrl *dependsController) emitChangeEvent(changeType string, spec PodSpec, pod Pod) {
}

func newDependsController(spec PodSpec, pods map[string]map[string]SharedPodWithSpec) *dependsController {
	depCtrl := &dependsController{
		Publisher: NewPublisher(true),
		spec:      spec,
		startedAt: time.Now(),
		podCtrls:  make(map[string]map[string]*sharedPodController),
		opsChan:   make(chan depOperation, 100),

		specStoredKey: strings.Join([]string{kLainDeploydRootKey, kLainDependencyKey, kLainSpecKey, spec.Name}, "/"),
		podsStoredKey: strings.Join([]string{kLainDeploydRootKey, kLainDependencyKey, kLainPodKey, spec.Name}, "/"),
	}

	for node, nsPods := range pods {
		for namespace, pod := range nsPods {
			podCtrl, _ := depCtrl.getOrAddPodCtrl(node, namespace, pod.Spec, pod.Pod)
			podCtrl.refCount = pod.RefCount
			podCtrl.verifyTime = pod.VerifyTime
		}
	}

	return depCtrl
}
