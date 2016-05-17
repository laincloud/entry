package engine

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/storage"
	"github.com/mijia/adoc"
	"github.com/mijia/sweb/log"
)

const (
	kDefaultRefreshInterval = 90
	kMinRefreshInterval     = 20
)

var (
	ErrPodGroupExists         = errors.New("PodGroup has already existed")
	ErrPodGroupNotExists      = errors.New("PodGroup not existed")
	ErrPodGroupCleaning       = errors.New("PodGroup is removing, need to wait for that")
	ErrNotEnoughResources     = errors.New("Not enough CPUs and Memory to use")
	ErrDependencyPodExists    = errors.New("DependencyPod has already existed")
	ErrDependencyPodNotExists = errors.New("DependencyPod not existed")
)

type OrcEngine struct {
	sync.RWMutex

	cluster      cluster.Cluster
	store        storage.Store
	eagleView    *RuntimeEagleView
	pgCtrls      map[string]*podGroupController
	rmPgCtrls    map[string]*podGroupController
	dependsCtrls map[string]*dependsController
	rmDepCtrls   map[string]*dependsController
	opsChan      chan orcOperation
	stop         chan struct{}
}

func (engine *OrcEngine) ListenerId() string {
	return "deployd.orc_engine"
}

func (engine *OrcEngine) HandleEvent(payload interface{}) {
	// Handle the dependency events
	if event, ok := payload.(DependencyEvent); ok {
		engine.RLock()
		defer engine.RUnlock()
		if depCtrl, ok := engine.dependsCtrls[event.Name]; ok {
			log.Debugf("Engine handle event %#v, dispatch it to dependsCtrl %s", event, depCtrl)
			engine.opsChan <- orcOperDependsDispatch{depCtrl, event}
		} else {
			// FIXME we are missing some dependency, should create the alarm
			log.Warnf("Engine found some missing dependency pod, %s", event.Name)
		}
		return
	}
}

func (engine *OrcEngine) NewDependencyPod(spec PodSpec) error {
	engine.Lock()
	defer engine.Unlock()

	if _, ok := engine.dependsCtrls[spec.Name]; ok {
		return ErrDependencyPodExists
	}
	if _, ok := engine.rmDepCtrls[spec.Name]; ok {
		return ErrDependencyPodExists
	}

	depCtrl := engine.initDependsCtrl(spec, nil)
	engine.dependsCtrls[spec.Name] = depCtrl
	engine.opsChan <- orcOperDependsAddSpec{depCtrl}
	return nil
}

func (engine *OrcEngine) GetDependencyPod(name string) (NamespacePodsWithSpec, error) {
	engine.RLock()
	defer engine.RUnlock()

	if depCtrl, ok := engine.dependsCtrls[name]; !ok {
		return NamespacePodsWithSpec{}, ErrDependencyPodNotExists
	} else {
		return depCtrl.Inspect(), nil
	}
}

func (engine *OrcEngine) UpdateDependencyPod(spec PodSpec) error {
	engine.RLock()
	defer engine.RUnlock()
	if depCtrl, ok := engine.dependsCtrls[spec.Name]; !ok {
		return ErrDependencyPodNotExists
	} else {
		engine.opsChan <- orcOperDependsUpdateSpec{depCtrl, spec}
		return nil
	}
}

func (engine *OrcEngine) RemoveDependencyPod(name string, force bool) error {
	engine.Lock()
	defer engine.Unlock()
	if depCtrl, ok := engine.dependsCtrls[name]; !ok {
		return ErrDependencyPodNotExists
	} else {
		engine.opsChan <- orcOperDependsRemoveSpec{depCtrl, force}
		delete(engine.dependsCtrls, name)
		engine.rmDepCtrls[name] = depCtrl
		go engine.checkDependsRemoveResult(name, depCtrl)
		return nil
	}
	return nil
}

func (engine *OrcEngine) GetNodes() ([]cluster.Node, error) {
	return engine.cluster.GetResources()
}

func (engine *OrcEngine) NewPodGroup(spec PodGroupSpec) error {
	engine.Lock()
	defer engine.Unlock()
	if _, ok := engine.pgCtrls[spec.Name]; ok {
		return ErrPodGroupExists
	}
	if _, ok := engine.rmPgCtrls[spec.Name]; ok {
		return ErrPodGroupCleaning
	}

	for _, depends := range spec.Pod.Dependencies {
		if _, ok := engine.dependsCtrls[depends.PodName]; !ok {
			//We will allow the weak reference to the dependency pods and won't return an error
			//FIXME: generate the alarm message or alarm data
			log.Warnf("Engine found some missing dependency pod, %s", depends.PodName)
		}
	}

	var pg PodGroup
	pg.State = RunStatePending
	pgCtrl := engine.initPodGroupCtrl(spec, nil, pg)
	engine.pgCtrls[spec.Name] = pgCtrl
	engine.opsChan <- orcOperDeploy{pgCtrl}
	return nil
}

func (engine *OrcEngine) InspectPodGroup(name string) (PodGroupWithSpec, bool) {
	engine.RLock()
	defer engine.RUnlock()
	if pgCtrl, ok := engine.pgCtrls[name]; !ok {
		return PodGroupWithSpec{}, false
	} else {
		return pgCtrl.Inspect(), true
	}
}

func (engine *OrcEngine) RefreshPodGroup(name string, forceUpdate bool) error {
	engine.RLock()
	defer engine.RUnlock()
	if pgCtrl, ok := engine.pgCtrls[name]; !ok {
		return ErrPodGroupNotExists
	} else {
		engine.opsChan <- orcOperRefresh{pgCtrl, forceUpdate}
		return nil
	}
}

func (engine *OrcEngine) RemovePodGroup(name string) error {
	engine.Lock()
	defer engine.Unlock()
	if pgCtrl, ok := engine.pgCtrls[name]; !ok {
		return ErrPodGroupNotExists
	} else {
		engine.opsChan <- orcOperRemove{pgCtrl}
		delete(engine.pgCtrls, name)
		engine.rmPgCtrls[name] = pgCtrl
		go engine.checkPodGroupRemoveResult(name, pgCtrl)
		return nil
	}
}

func (engine *OrcEngine) RescheduleInstance(name string, numInstances int, restartPolicy ...RestartPolicy) error {
	engine.RLock()
	defer engine.RUnlock()
	if pgCtrl, ok := engine.pgCtrls[name]; !ok {
		return ErrPodGroupNotExists
	} else {
		engine.opsChan <- orcOperRescheduleInstance{pgCtrl, numInstances, restartPolicy}
		return nil
	}
}

func (engine *OrcEngine) RescheduleSpec(name string, podSpec PodSpec) error {
	engine.RLock()
	defer engine.RUnlock()
	if pgCtrl, ok := engine.pgCtrls[name]; !ok {
		return ErrPodGroupNotExists
	} else {
		for _, depends := range podSpec.Dependencies {
			if _, ok := engine.dependsCtrls[depends.PodName]; !ok {
				// We will allow the weak reference to the dependency pods and won't return an error
				// FIXME: generate the alarm message or alarm data
				log.Warnf("Engine found some missing dependency pod, %s", depends.PodName)
			}
		}
		engine.opsChan <- orcOperRescheduleSpec{pgCtrl, podSpec}
		return nil
	}
}

func (engine *OrcEngine) Start() {
	engine.Lock()
	defer engine.Unlock()
	log.Infof("Start engine...")
	if engine.stop != nil { // stop is not nil, means it having been started
		log.Debugf("Engine having been started, ignore.")
		return
	}
	engine.stop = make(chan struct{})
	go engine.initOperationWorker()
	go engine.startClusterMonitor()
}

func (engine *OrcEngine) Stop() {
	engine.Lock()
	defer engine.Unlock()
	log.Infof("Stop engine...")
	if engine.stop == nil {
		log.Debugf("Engine having been stop, ignore.")
		return
	}
	select {
	case _, ok := <-engine.stop:
		if !ok {
			return // channel was closed
		}
	default:
	}
	close(engine.stop)
	engine.stop = nil
}

func (engine *OrcEngine) Started() bool {
	return engine.stop != nil
}

func (engine *OrcEngine) LoadDependsPods() error {
	depCtrls := make(map[string]*dependsController)
	specDirKey := strings.Join([]string{kLainDeploydRootKey, kLainDependencyKey, kLainSpecKey}, "/")
	if specNames, err := engine.store.KeysByPrefix(specDirKey); err != nil {
		if err != storage.ErrNoSuchKey {
			return err
		}
	} else {
		for _, name := range specNames {
			var spec PodSpec
			if err := engine.store.Get(name, &spec); err != nil {
				log.Errorf("Failed to load dependency pod spec %q from storage, %s", name, err)
				return err
			}

			var pods map[string]map[string]SharedPodWithSpec
			podsKey := strings.Join([]string{kLainDeploydRootKey, kLainDependencyKey, kLainPodKey, spec.Name}, "/")
			if err := engine.store.Get(podsKey, &pods); err != nil {
				if err != storage.ErrNoSuchKey {
					log.Errorf("Failed to load dependency pods runtime %q from storage, %s", podsKey, err)
					return err
				} else {
					// we should allow only have the spec but no pods
					log.Warnf("Found empty dependency pods runtime %q from storage, %s", podsKey, err)
				}
			}
			depCtrls[spec.Name] = engine.initDependsCtrl(spec, pods)
			log.Infof("Loaded DependsController, %s", depCtrls[spec.Name])
		}
	}
	engine.dependsCtrls = depCtrls
	return nil
}

func (engine *OrcEngine) LoadPodGroups() error {
	pgCtrls := make(map[string]*podGroupController)
	pgKey := fmt.Sprintf("%s/%s", kLainDeploydRootKey, kLainPodGroupKey)
	if pgNamespaces, err := engine.store.KeysByPrefix(pgKey); err != nil {
		if err != storage.ErrNoSuchKey {
			return err
		}
	} else {
		for _, pgNamespace := range pgNamespaces {
			pgNames, err := engine.store.KeysByPrefix(pgNamespace)
			if err != nil {
				if err != storage.ErrNoSuchKey {
					return err
				}
			}
			for _, pgName := range pgNames {
				var pgWithSpec PodGroupWithSpec
				if err := engine.store.Get(pgName, &pgWithSpec); err != nil {
					log.Errorf("Failed to load pod group with spec %q from storage, %s", pgName, err)
					return err
				}
				spec, states, pg := pgWithSpec.Spec, pgWithSpec.PrevState, pgWithSpec.PodGroup
				pgCtrls[spec.Name] = engine.initPodGroupCtrl(spec, states, pg)
				log.Infof("Loaded PodGroupController, %s", pgCtrls[spec.Name])
			}
		}
	}
	engine.pgCtrls = pgCtrls
	return nil
}

func (engine *OrcEngine) initDependsCtrl(spec PodSpec, pods map[string]map[string]SharedPodWithSpec) *dependsController {
	depCtrl := newDependsController(spec, pods)
	depCtrl.Activate(engine.cluster, engine.store, engine.eagleView, engine.stop)
	return depCtrl
}

func (engine *OrcEngine) initPodGroupCtrl(spec PodGroupSpec, states []PodPrevState, pg PodGroup) *podGroupController {
	pgCtrl := newPodGroupController(spec, states, pg)
	pgCtrl.AddListener(engine)
	pgCtrl.Activate(engine.cluster, engine.store, engine.eagleView, engine.stop)
	return pgCtrl
}

// This will be running inside the go routine
func (engine *OrcEngine) initOperationWorker() {
	tick := time.Tick(kDefaultRefreshInterval * time.Second)
	for {
		select {
		case op := <-engine.opsChan:
			op.Do(engine)
		case <-tick:
			engine.RLock()
			if len(engine.pgCtrls) > 0 {
				rInterval := kDefaultRefreshInterval / 2 * 1000 / len(engine.pgCtrls)
				index := 0
				for _, pgCtrl := range engine.pgCtrls {
					interval := index * rInterval
					_pgCtrl := pgCtrl
					index++
					go func() {
						log.Infof("%s will be refreshed after %d seconds", _pgCtrl, interval/1000)
						time.Sleep(time.Duration(interval) * time.Millisecond)
						engine.opsChan <- orcOperRefresh{_pgCtrl, false}
					}()
				}
			}
			if len(engine.dependsCtrls) > 0 {
				rInterval := kDefaultRefreshInterval / 2 * 1000 / len(engine.dependsCtrls)
				index := 0
				for _, depCtrl := range engine.dependsCtrls {
					interval := index * rInterval
					_depCtrl := depCtrl
					index++
					go func() {
						log.Infof("%s will be refreshed after %d seconds", _depCtrl, interval/1000)
						time.Sleep(time.Duration(interval) * time.Millisecond)
						engine.opsChan <- orcOperDependsRefresh{_depCtrl}
					}()
				}
			}
			engine.RUnlock()
		case <-engine.stop:
			if len(engine.opsChan) == 0 {
				return
			}
		}
	}
}

// This will be running inside the go routine
func (engine *OrcEngine) checkDependsRemoveResult(name string, depCtrl *dependsController) {
	tick := time.Tick(5 * time.Second)
	for _ = range tick {
		switch depCtrl.RemoveStatus() {
		case 1:
			log.Infof("<OrcEngine> DependsCtrl %s is safely removed", name)
			engine.Lock()
			delete(engine.rmDepCtrls, name)
			engine.Unlock()
			return
		case 2:
			log.Infof("<OrcEngine> DependsCtrl %s cannot be removed, someone maybe using it", name)
			engine.Lock()
			engine.dependsCtrls[name] = depCtrl
			delete(engine.rmDepCtrls, name)
			engine.Unlock()
			return
		}
	}
}

// This will be running inside the go routine
func (engine *OrcEngine) checkPodGroupRemoveResult(name string, pgCtrl *podGroupController) {
	timeout := time.After(60 * time.Second)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-tick:
			if pgCtrl.IsRemoved() {
				log.Infof("<OrcEngine> PodGroup %s is safely removed", name)
				engine.Lock()
				delete(engine.rmPgCtrls, name)
				engine.Unlock()
				return
			}
		case <-timeout:
			log.Errorf("!!!<OrcEngine> timeout when checking pod group results, pg %s need to be checked and removed annually.", name)
			engine.Lock()
			delete(engine.rmPgCtrls, name)
			engine.Unlock()
			return
		}
	}
}

func (engine *OrcEngine) DriftNode(fromNode, toNode string, pgName string, pgInstance int, force bool) {
	engine.RLock()
	defer engine.RUnlock()
	if pgName == "" {
		for _, pgCtrl := range engine.pgCtrls {
			_pgCtrl := pgCtrl
			engine.opsChan <- orcOperScheduleDrift{_pgCtrl, fromNode, toNode, pgInstance, force}
		}
	} else {
		if pgCtrl, ok := engine.pgCtrls[pgName]; ok {
			engine.opsChan <- orcOperScheduleDrift{pgCtrl, fromNode, toNode, pgInstance, force}
		}
	}
	// FIXME: do we need to tell dependsCtrl to drift?
	// so far we just wait for the dependsCtrl to react to the events
}

func (engine *OrcEngine) onClusterNodeLost(nodeName string) {
	log.Warnf("Cluster node is down, [%q], will notify pod group controller to drift", nodeName)
	engine.DriftNode(nodeName, "", "", -1, false)
}

func (engine *OrcEngine) startClusterMonitor() {
	restart := make(chan bool)
	eventMonitorId := engine.cluster.MonitorEvents("", func(event adoc.Event, err error) {
		if err != nil {
			log.Warnf("Error during the cluster event monitor, will try to restart the monitor, %s", err)
			restart <- true
		} else {
			log.Debugf("Cluster event: %+v", event)
			if strings.HasPrefix(event.From, "swarm") {
				switch event.Status {
				case "node_disconnect":
					engine.onClusterNodeLost(event.Node.Name)
				}
			}
		}
	})
	select {
	case <-restart:
		engine.cluster.StopMonitor(eventMonitorId)
		close(restart)
		time.Sleep(200 * time.Millisecond)
		engine.startClusterMonitor()
	case <-engine.stop:
		engine.cluster.StopMonitor(eventMonitorId)
	}
}

func New(cluster cluster.Cluster, store storage.Store) (*OrcEngine, error) {
	engine := &OrcEngine{
		cluster:      cluster,
		store:        store,
		pgCtrls:      make(map[string]*podGroupController),
		rmPgCtrls:    make(map[string]*podGroupController),
		dependsCtrls: make(map[string]*dependsController),
		rmDepCtrls:   make(map[string]*dependsController),
		opsChan:      make(chan orcOperation, 500),
		stop:         nil,
	}

	eagleView := NewRuntimeEagleView()
	//if err := eagleView.Refresh(cluster); err != nil {
	//log.Warnf("<OrcEngine> Cannot refresh all the runtime data for bootstraping, %s", err)
	//return nil, err
	//}
	engine.eagleView = eagleView

	if err := engine.LoadDependsPods(); err != nil {
		return nil, err
	}

	if err := engine.LoadPodGroups(); err != nil {
		return nil, err
	}

	engine.Start()

	return engine, nil
}
