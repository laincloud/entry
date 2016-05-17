package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/storage"
	"github.com/mijia/sweb/log"
)

type PodGroupWithSpec struct {
	Spec      PodGroupSpec
	PrevState []PodPrevState
	PodGroup
}

type podGroupController struct {
	Publisher

	sync.RWMutex
	spec      PodGroupSpec
	prevState []PodPrevState
	group     PodGroup

	evSnapshot []RuntimeEaglePod
	podCtrls   []*podController
	opsChan    chan pgOperation

	storedKey    string
	storedKeyDir string
	infoKey      string
	infoKeyDir   string
}

func (pgCtrl *podGroupController) String() string {
	return fmt.Sprintf("PodGroupCtrl %s", pgCtrl.spec)
}

func (pgCtrl *podGroupController) Inspect() PodGroupWithSpec {
	pgCtrl.RLock()
	defer pgCtrl.RUnlock()
	return PodGroupWithSpec{pgCtrl.spec, pgCtrl.prevState, pgCtrl.group}
}

func (pgCtrl *podGroupController) IsRemoved() bool {
	pgCtrl.RLock()
	defer pgCtrl.RUnlock()
	return pgCtrl.group.State == RunStateRemoved
}

func (pgCtrl *podGroupController) IsPending() bool {
	pgCtrl.RLock()
	defer pgCtrl.RUnlock()
	return pgCtrl.group.State == RunStatePending
}

func (pgCtrl *podGroupController) Deploy() {
	pgCtrl.RLock()
	spec := pgCtrl.spec.Clone()
	pgCtrl.RUnlock()

	pgCtrl.opsChan <- pgOperLogOperation{"Start to deploy"}
	pgCtrl.opsChan <- pgOperSaveStore{true}
	pgCtrl.opsChan <- pgOperSnapshotEagleView{spec.Name}
	for i := 0; i < spec.NumInstances; i += 1 {
		pgCtrl.opsChan <- pgOperDeployInstance{i + 1, spec.Version}
	}
	pgCtrl.opsChan <- pgOperSnapshotGroup{true}
	pgCtrl.opsChan <- pgOperSnapshotPrevState{}
	pgCtrl.opsChan <- pgOperSaveStore{true}
	pgCtrl.opsChan <- pgOperLogOperation{"deploy finished"}
}

func (pgCtrl *podGroupController) RescheduleInstance(numInstances int, restartPolicy ...RestartPolicy) {
	pgCtrl.RLock()
	spec := pgCtrl.spec.Clone()
	pgCtrl.RUnlock()

	isDirty := false
	curNumInstances := spec.NumInstances
	if numInstances >= 0 && curNumInstances != numInstances {
		spec.NumInstances = numInstances
		isDirty = true
	}
	if len(restartPolicy) > 0 && pgCtrl.spec.RestartPolicy != restartPolicy[0] {
		spec.RestartPolicy = restartPolicy[0]
		isDirty = true
	}
	if !isDirty {
		return
	}

	pgCtrl.Lock()
	pgCtrl.spec = spec
	pgCtrl.Unlock()
	pgCtrl.opsChan <- pgOperLogOperation{fmt.Sprintf("Start to reschedule instance from %d to %d", curNumInstances, numInstances)}
	pgCtrl.opsChan <- pgOperSaveStore{true}
	delta := numInstances - curNumInstances
	if delta != 0 {
		pgCtrl.opsChan <- pgOperSnapshotEagleView{spec.Name}
		if delta > 0 {
			for i := 0; i < delta; i += 1 {
				instanceNo := i + 1 + curNumInstances
				pgCtrl.opsChan <- pgOperPushPodCtrl{spec.Pod}
				pgCtrl.opsChan <- pgOperDeployInstance{instanceNo, spec.Version}
			}
		} else {
			delta *= -1
			for i := 0; i < delta; i += 1 {
				pgCtrl.opsChan <- pgOperRemoveInstance{curNumInstances - i, spec.Pod}
				pgCtrl.opsChan <- pgOperPopPodCtrl{}
			}
		}
		pgCtrl.opsChan <- pgOperSnapshotGroup{true}
		pgCtrl.opsChan <- pgOperSnapshotPrevState{}
		pgCtrl.opsChan <- pgOperSaveStore{true}
	}
	pgCtrl.opsChan <- pgOperLogOperation{"Reschedule instance number finished"}
}

func (pgCtrl *podGroupController) RescheduleSpec(podSpec PodSpec) {
	pgCtrl.RLock()
	spec := pgCtrl.spec.Clone()
	pgCtrl.RUnlock()

	if spec.Pod.Equals(podSpec) {
		return
	}

	oldPodSpec := spec.Pod.Clone()
	spec.Pod = spec.Pod.Merge(podSpec)
	spec.Version += 1
	spec.UpdatedAt = time.Now()
	pgCtrl.Lock()
	pgCtrl.spec = spec
	pgCtrl.Unlock()
	pgCtrl.opsChan <- pgOperLogOperation{"Start to reschedule spec"}
	pgCtrl.opsChan <- pgOperSaveStore{true}
	pgCtrl.opsChan <- pgOperSnapshotEagleView{spec.Name}

	for i := 0; i < spec.NumInstances; i += 1 {
		pgCtrl.opsChan <- pgOperUpgradeInstance{i + 1, spec.Version, oldPodSpec, spec.Pod}
		// wait some seconds for new instance's initialization completed, before we update next one
		time.Sleep(time.Second * time.Duration(podSpec.GetSetupTime()))
	}
	pgCtrl.opsChan <- pgOperSnapshotGroup{true}
	pgCtrl.opsChan <- pgOperSnapshotPrevState{}
	pgCtrl.opsChan <- pgOperSaveStore{true}
	pgCtrl.opsChan <- pgOperLogOperation{"Reschedule spec finished"}
}

func (pgCtrl *podGroupController) RescheduleDrift(fromNode, toNode string, instanceNo int, force bool) {
	pgCtrl.RLock()
	spec := pgCtrl.spec.Clone()
	pgCtrl.RUnlock()

	if spec.NumInstances == 0 {
		return
	}

	pgCtrl.opsChan <- pgOperLogOperation{fmt.Sprintf("Start to reschedule drift from %s", fromNode)}
	if instanceNo == -1 {
		for i := 0; i < spec.NumInstances; i += 1 {
			pgCtrl.opsChan <- pgOperDriftInstance{i + 1, fromNode, toNode, force}
		}
	} else {
		pgCtrl.opsChan <- pgOperDriftInstance{instanceNo, fromNode, toNode, force}
	}
	pgCtrl.opsChan <- pgOperSnapshotGroup{false}
	pgCtrl.opsChan <- pgOperSnapshotPrevState{}
	pgCtrl.opsChan <- pgOperSaveStore{false}
	pgCtrl.opsChan <- pgOperLogOperation{"Reschedule drift finished"}
}

func (pgCtrl *podGroupController) Remove() {
	pgCtrl.RLock()
	spec := pgCtrl.spec.Clone()
	pgCtrl.RUnlock()

	pgCtrl.opsChan <- pgOperLogOperation{"Start to remove"}
	pgCtrl.opsChan <- pgOperRemoveStore{}
	for i := 0; i < spec.NumInstances; i += 1 {
		pgCtrl.opsChan <- pgOperRemoveInstance{i + 1, spec.Pod}
	}
	pgCtrl.opsChan <- pgOperLogOperation{"Remove finished"}
	pgCtrl.opsChan <- pgOperSnapshotEagleView{spec.Name}
	pgCtrl.opsChan <- pgOperPurge{}
}

func (pgCtrl *podGroupController) Refresh(force bool) {
	if pgCtrl.IsRemoved() || pgCtrl.IsPending() {
		return
	}

	pgCtrl.RLock()
	spec := pgCtrl.spec.Clone()
	pgCtrl.RUnlock()

	pgCtrl.opsChan <- pgOperLogOperation{"Start to refresh PodGroup"}
	pgCtrl.opsChan <- pgOperSnapshotEagleView{spec.Name}
	for i := 0; i < spec.NumInstances; i += 1 {
		pgCtrl.opsChan <- pgOperRefreshInstance{i + 1, spec}
	}
	pgCtrl.opsChan <- pgOperVerifyInstanceCount{spec}
	pgCtrl.opsChan <- pgOperSnapshotGroup{force}
	pgCtrl.opsChan <- pgOperSnapshotPrevState{}
	pgCtrl.opsChan <- pgOperSaveStore{false}
	pgCtrl.opsChan <- pgOperLogOperation{"PodGroup refreshing finished"}
}

func (pgCtrl *podGroupController) Activate(c cluster.Cluster, store storage.Store, eagle *RuntimeEagleView, stop chan struct{}) {
	go func() {
		for {
			select {
			case op := <-pgCtrl.opsChan:
				toShutdown := op.Do(pgCtrl, c, store, eagle)
				if toShutdown {
					return
				}
			case <-stop:
				if len(pgCtrl.opsChan) == 0 {
					return
				}
			}
		}
	}()
}

func (pgCtrl *podGroupController) emitChangeEvent(changeType string, spec PodSpec, pod Pod, nodeName string) {
	if changeType == "" || nodeName == "" {
		return
	}
	var events []interface{}
	namespace := spec.Namespace
	for _, dep := range spec.Dependencies {
		if dep.Policy == DependencyNodeLevel {
			namespace = fmt.Sprintf("%s", nodeName)
		}
		events = append(events, DependencyEvent{
			Type:      changeType,
			Name:      dep.PodName,
			NodeName:  nodeName,
			Namespace: namespace,
		})
	}
	log.Debugf("%s emit change event: %s, %q, #evts=%d", pgCtrl, changeType, nodeName, len(events))
	for _, evt := range events {
		pgCtrl.EmitEvent(evt)
	}
}

func newPodGroupController(spec PodGroupSpec, states []PodPrevState, pg PodGroup) *podGroupController {
	podCtrls := make([]*podController, spec.NumInstances)
	for i := range podCtrls {
		var pod Pod
		pod.InstanceNo = i + 1
		pod.State = RunStatePending
		podSpec := spec.Pod.Clone()
		if states != nil && i < len(states) {
			podSpec.PrevState = states[i].Clone() // set the pod's prev state
		} else {
			podSpec.PrevState = NewPodPrevState(1) // set empty prev state
		}
		podCtrls[i] = &podController{
			spec: podSpec,
			pod:  pod,
		}
	}
	// we may have some running pods loading from the storage
	for _, pod := range pg.Pods {
		if pod.InstanceNo < 1 || pod.InstanceNo > spec.NumInstances {
			log.Warnf("We have some pod have InstanceNo out of bounds, %d, bounds=[0, %d]", pod.InstanceNo, spec.NumInstances)
			continue
		}
		podCtrls[pod.InstanceNo-1].pod = pod
	}

	pgCtrl := &podGroupController{
		spec:     spec,
		group:    pg,
		podCtrls: podCtrls,
		opsChan:  make(chan pgOperation, 500),

		storedKey:    strings.Join([]string{kLainDeploydRootKey, kLainPodGroupKey, spec.Namespace, spec.Name}, "/"),
		storedKeyDir: strings.Join([]string{kLainDeploydRootKey, kLainPodGroupKey, spec.Namespace}, "/"),
	}
	pgCtrl.Publisher = NewPublisher(true)
	return pgCtrl
}
