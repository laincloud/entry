package engine

type orcOperation interface {
	Do(engine *OrcEngine)
}

// Depends Operations

type orcOperDependsAddSpec struct {
	depCtrl *dependsController
}

func (op orcOperDependsAddSpec) Do(engine *OrcEngine) {
	op.depCtrl.AddSpec()
}

type orcOperDependsUpdateSpec struct {
	depCtrl *dependsController
	newSpec PodSpec
}

func (op orcOperDependsUpdateSpec) Do(engine *OrcEngine) {
	op.depCtrl.UpdateSpec(op.newSpec)
}

type orcOperDependsRemoveSpec struct {
	depCtrl *dependsController
	force   bool
}

func (op orcOperDependsRemoveSpec) Do(engine *OrcEngine) {
	op.depCtrl.RemoveSpec(op.force)
}

type orcOperDependsRefresh struct {
	depCtrl *dependsController
}

func (op orcOperDependsRefresh) Do(engine *OrcEngine) {
	op.depCtrl.Refresh()
}

type orcOperDependsDispatch struct {
	depCtrl *dependsController
	event   DependencyEvent
}

func (op orcOperDependsDispatch) Do(engine *OrcEngine) {
	event := op.event
	switch event.Type {
	case "add":
		op.depCtrl.AddPod(event.Namespace, event.NodeName)
	case "remove":
		op.depCtrl.RemovePod(event.Namespace, event.NodeName)
	case "verify":
		op.depCtrl.VerifyPod(event.Namespace, event.NodeName)
	}
}

// PodGroup Operations

type orcOperDeploy struct {
	pgCtrl *podGroupController
}

func (op orcOperDeploy) Do(engine *OrcEngine) {
	op.pgCtrl.Deploy()
}

type orcOperRefresh struct {
	pgCtrl      *podGroupController
	forceUpdate bool
}

func (op orcOperRefresh) Do(engine *OrcEngine) {
	op.pgCtrl.Refresh(op.forceUpdate)
}

type orcOperRemove struct {
	pgCtrl *podGroupController
}

func (op orcOperRemove) Do(engine *OrcEngine) {
	op.pgCtrl.Remove()
}

type orcOperRescheduleInstance struct {
	pgCtrl        *podGroupController
	numInstances  int
	restartPolicy []RestartPolicy
}

func (op orcOperRescheduleInstance) Do(engine *OrcEngine) {
	op.pgCtrl.RescheduleInstance(op.numInstances, op.restartPolicy...)
}

type orcOperRescheduleSpec struct {
	pgCtrl  *podGroupController
	podSpec PodSpec
}

func (op orcOperRescheduleSpec) Do(engine *OrcEngine) {
	op.pgCtrl.RescheduleSpec(op.podSpec)
}

type orcOperScheduleDrift struct {
	pgCtrl     *podGroupController
	fromNode   string
	toNode     string
	instanceNo int
	force      bool
}

func (op orcOperScheduleDrift) Do(engine *OrcEngine) {
	op.pgCtrl.RescheduleDrift(op.fromNode, op.toNode, op.instanceNo, op.force)
}
