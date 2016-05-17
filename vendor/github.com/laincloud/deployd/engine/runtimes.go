package engine

import (
	"time"

	"github.com/mijia/adoc"
)

type RunState int

const (
	RunStatePending = iota
	RunStateDrift
	RunStateSuccess
	RunStateExit
	RunStateFail
	RunStateInconsistent
	RunStateMissing
	RunStateRemoved
)

func (rs RunState) String() string {
	switch rs {
	case RunStatePending:
		return "RunStatePending"
	case RunStateDrift:
		return "RunStateDrift"
	case RunStateSuccess:
		return "RunStateSuccess"
	case RunStateExit:
		return "RunStateExit"
	case RunStateFail:
		return "RunStateFail"
	case RunStateMissing:
		return "RunStateMissing"
	case RunStateInconsistent:
		return "RunStateInconsistent"
	case RunStateRemoved:
		return "RunStateRemoved"
	default:
		return "Unknown RunState"
	}
}

type ImRuntime struct {
	State      RunState
	LastError  string
	DriftCount int
	UpdatedAt  time.Time
}

type Container struct {
	// FIXME(mijia): multiple ports supporing, will have multiple entries of <NodePort, ContainerPort, Protocol>
	Id            string
	Runtime       adoc.ContainerDetail
	NodeName      string
	NodeIp        string
	ContainerIp   string
	NodePort      int
	ContainerPort int
	Protocol      string
}

func (c Container) Clone() Container {
	// So far we maybe only care about the basic information like in the Equals
	return c
}

func (c Container) Equals(o Container) bool {
	// The ContainerDetail from adoc change would reflect to the Pod runtime changes
	return c.Id == o.Id &&
		c.NodeName == o.NodeName &&
		c.NodeIp == o.NodeIp &&
		c.ContainerIp == o.ContainerIp &&
		c.NodePort == o.NodePort &&
		c.ContainerPort == o.ContainerPort &&
		c.Protocol == o.Protocol
}

type Pod struct {
	InstanceNo int
	Containers []Container
	ImRuntime
}

func (p Pod) Clone() Pod {
	n := p
	n.Containers = make([]Container, len(p.Containers))
	for i := range p.Containers {
		n.Containers[i] = p.Containers[i].Clone()
	}
	return n
}

func (p Pod) Equals(o Pod) bool {
	if len(p.Containers) != len(o.Containers) {
		return false
	}
	for i := range p.Containers {
		if !p.Containers[i].Equals(o.Containers[i]) {
			return false
		}
	}
	return p.InstanceNo == o.InstanceNo &&
		p.State == o.State &&
		p.LastError == o.LastError &&
		p.DriftCount == o.DriftCount
}

func (pod Pod) ContainerIds() []string {
	ids := make([]string, len(pod.Containers))
	for i, container := range pod.Containers {
		ids[i] = container.Id
	}
	return ids
}

func (pod Pod) NeedRestart(policy RestartPolicy) bool {
	state := pod.State
	if policy == RestartPolicyAlways {
		return state == RunStateExit || state == RunStateFail
	}
	if policy == RestartPolicyOnFail {
		return state == RunStateFail
	}
	return false
}

func (pod Pod) NodeName() string {
	if len(pod.Containers) > 0 {
		return pod.Containers[0].NodeName
	}
	return ""
}

func (pod Pod) NodeIp() string {
	if len(pod.Containers) > 0 {
		return pod.Containers[0].NodeIp
	}
	return ""
}

type PodGroup struct {
	Pods []Pod
	ImRuntime
}

func (pg PodGroup) Clone() PodGroup {
	n := pg
	n.Pods = make([]Pod, len(pg.Pods))
	for i := range pg.Pods {
		n.Pods[i] = pg.Pods[i].Clone()
	}
	return n
}

func (pg PodGroup) Equals(o PodGroup) bool {
	if len(pg.Pods) != len(o.Pods) {
		return false
	}
	for i := range pg.Pods {
		if !pg.Pods[i].Equals(o.Pods[i]) {
			return false
		}
	}
	return pg.State == o.State &&
		pg.LastError == o.LastError &&
		pg.DriftCount == o.DriftCount
}

func (group PodGroup) collectNodes() map[string]string {
	nodes := make(map[string]string)
	for _, pod := range group.Pods {
		name := pod.NodeName()
		ip := pod.NodeIp()
		if name != "" && ip != "" {
			nodes[name] = ip
		}
	}
	return nodes
}

type DependencyEvent struct {
	Type      string // add, remove, verify
	Name      string
	NodeName  string
	Namespace string
}
