package engine

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mijia/adoc"
	"github.com/mijia/go-generics"
)

const (
	kLainDeploydRootKey = "/lain/deployd"
	kLainPodGroupKey    = "pod_groups"
	kLainDependencyKey  = "depends"
	kLainSpecKey        = "specs"
	kLainPodKey         = "pods"
	kLainNodesKey       = "nodes"

	kLainVolumeRoot  = "/data/lain/volumes"
	kLainLabelPrefix = "cc.bdp.lain.deployd"

	MinPodSetupTime = 0
	MaxPodSetupTime = 300

	MinPodKillTimeout = 10
	MaxPodKillTimeout = 120
)

type ImSpec struct {
	Name      string
	Namespace string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ContainerLabel struct {
	Name           string
	Namespace      string
	InstanceNo     int
	Version        int
	DriftCount     int
	ContainerIndex int
	Annotation     string
}

func (label ContainerLabel) NameAffnity() string {
	return fmt.Sprintf("affnity:%s.pg_name!=~%s", kLainLabelPrefix, label.Name)
}

func (label ContainerLabel) Label2Maps() map[string]string {
	labelMaps := make(map[string]string)
	labelMaps[kLainLabelPrefix+".pg_name"] = label.Name
	labelMaps[kLainLabelPrefix+".pg_namespace"] = label.Namespace
	labelMaps[kLainLabelPrefix+".instance_no"] = fmt.Sprintf("%d", label.InstanceNo)
	labelMaps[kLainLabelPrefix+".version"] = fmt.Sprintf("%d", label.Version)
	labelMaps[kLainLabelPrefix+".drift_count"] = fmt.Sprintf("%d", label.DriftCount)
	labelMaps[kLainLabelPrefix+".container_index"] = fmt.Sprintf("%d", label.ContainerIndex)
	labelMaps[kLainLabelPrefix+".annotation"] = label.Annotation
	return labelMaps
}

func (label *ContainerLabel) FromMaps(m map[string]string) bool {
	var err error
	hasError := false
	label.Name = m[kLainLabelPrefix+".pg_name"]
	hasError = hasError || label.Name == ""
	label.Namespace = m[kLainLabelPrefix+".pg_namespace"]
	label.InstanceNo, err = strconv.Atoi(m[kLainLabelPrefix+".instance_no"])
	hasError = hasError || err != nil
	label.Version, err = strconv.Atoi(m[kLainLabelPrefix+".version"])
	hasError = hasError || err != nil
	label.DriftCount, err = strconv.Atoi(m[kLainLabelPrefix+".drift_count"])
	hasError = hasError || err != nil
	label.ContainerIndex, err = strconv.Atoi(m[kLainLabelPrefix+".container_index"])
	hasError = hasError || err != nil
	label.Annotation = m[kLainLabelPrefix+".annotation"]
	return !hasError
}

type ContainerSpec struct {
	ImSpec
	Image         string
	Env           []string
	User          string
	WorkingDir    string
	DnsSearch     []string
	Volumes       []string // a stateful flag
	SystemVolumes []string // not a stateful flag, every node has system volumes
	Command       []string
	Entrypoint    []string
	CpuLimit      int
	MemoryLimit   int64
	Expose        int
	LogConfig     adoc.LogConfig
}

func (s ContainerSpec) Clone() ContainerSpec {
	newSpec := s
	newSpec.Env = generics.Clone_StringSlice(s.Env)
	newSpec.Volumes = generics.Clone_StringSlice(s.Volumes)
	newSpec.SystemVolumes = generics.Clone_StringSlice(s.SystemVolumes)
	newSpec.Command = generics.Clone_StringSlice(s.Command)
	newSpec.DnsSearch = generics.Clone_StringSlice(s.DnsSearch)
	newSpec.Entrypoint = generics.Clone_StringSlice(s.Entrypoint)
	newSpec.LogConfig.Type = s.LogConfig.Type
	newSpec.LogConfig.Config = generics.Clone_StringStringMap(s.LogConfig.Config)
	return newSpec
}

func (s ContainerSpec) VerifyParams() bool {
	verify := s.Image != "" &&
		s.CpuLimit >= 0 &&
		s.MemoryLimit >= 0 &&
		s.Expose >= 0

	return verify
}

func (s ContainerSpec) Equals(o ContainerSpec) bool {
	return s.Name == o.Name &&
		s.Image == o.Image &&
		generics.Equal_StringSlice(s.Env, o.Env) &&
		generics.Equal_StringSlice(s.Command, o.Command) &&
		generics.Equal_StringSlice(s.DnsSearch, o.DnsSearch) &&
		s.CpuLimit == o.CpuLimit &&
		s.MemoryLimit == o.MemoryLimit &&
		s.Expose == o.Expose &&
		s.User == o.User &&
		s.WorkingDir == o.WorkingDir &&
		generics.Equal_StringSlice(s.Volumes, o.Volumes) &&
		generics.Equal_StringSlice(s.SystemVolumes, o.SystemVolumes) &&
		generics.Equal_StringSlice(s.Entrypoint, o.Entrypoint) &&
		s.LogConfig.Type == o.LogConfig.Type &&
		generics.Equal_StringStringMap(s.LogConfig.Config, o.LogConfig.Config)
}

func NewContainerSpec(image string) ContainerSpec {
	spec := ContainerSpec{
		Image: image,
	}
	spec.Version = 1
	spec.CreatedAt = time.Now()
	spec.UpdatedAt = spec.CreatedAt
	return spec
}

type DependencyPolicy int

const (
	DependencyNamespaceLevel = iota
	DependencyNodeLevel
)

type Dependency struct {
	PodName string
	Policy  DependencyPolicy
}

func (d Dependency) Clone() Dependency {
	return d
}

type PodPrevState struct {
	NodeName string
	IPs      []string
}

func NewPodPrevState(length int) PodPrevState {
	return PodPrevState{
		NodeName: "",
		IPs:      make([]string, length),
	}
}

func (pps PodPrevState) Clone() PodPrevState {
	newState := pps
	newState.IPs = make([]string, len(pps.IPs))
	copy(newState.IPs, pps.IPs)
	return newState
}

type PodSpec struct {
	ImSpec
	Containers   []ContainerSpec
	Filters      []string // for cluster scheduling
	Dependencies []Dependency
	Annotation   string
	Stateful     bool
	SetupTime    int
	KillTimeout  int
	PrevState    PodPrevState
}

func (s PodSpec) GetSetupTime() int {
	if s.SetupTime < MinPodSetupTime {
		return MinPodSetupTime
	} else if s.SetupTime > MaxPodSetupTime {
		return MaxPodSetupTime
	}
	return s.SetupTime
}

func (s PodSpec) GetKillTimeout() int {
	if s.KillTimeout < MinPodKillTimeout {
		return MinPodKillTimeout
	} else if s.KillTimeout > MaxPodKillTimeout {
		return MaxPodKillTimeout
	}
	return s.KillTimeout
}

func (s PodSpec) String() string {
	return fmt.Sprintf("Pod[name=%s, version=%d, depends=%+v, stateful=%v, #containers=%d]",
		s.Name, s.Version, s.Dependencies, s.Stateful, len(s.Containers))
}

func (s PodSpec) Clone() PodSpec {
	newSpec := s
	newSpec.Filters = generics.Clone_StringSlice(s.Filters)
	newSpec.Containers = make([]ContainerSpec, len(s.Containers))
	newSpec.PrevState = s.PrevState.Clone()
	for i := range s.Containers {
		newSpec.Containers[i] = s.Containers[i].Clone()
	}
	newSpec.Dependencies = make([]Dependency, len(s.Dependencies))
	for i := range s.Dependencies {
		newSpec.Dependencies[i] = s.Dependencies[i].Clone()
	}
	return newSpec
}

func (s PodSpec) VerifyParams() bool {
	verify := s.Name != "" && s.Namespace != "" &&
		len(s.Containers) > 0
	if !verify {
		return false
	}
	for _, cSpec := range s.Containers {
		if !cSpec.VerifyParams() {
			return false
		}
	}
	return true
}

func (s PodSpec) IsHardStateful() bool {
	return s.Stateful
}

func (s PodSpec) IsStateful() bool {
	return s.HasVolumes() || s.Stateful
}

func (s PodSpec) HasVolumes() bool {
	for _, container := range s.Containers {
		if len(container.Volumes) > 0 {
			return true
		}
	}
	return false
}

func (s PodSpec) Equals(o PodSpec) bool {
	if len(s.Containers) != len(o.Containers) {
		return false
	}
	for i := range s.Containers {
		if !s.Containers[i].Equals(o.Containers[i]) {
			return false
		}
	}
	if len(s.Dependencies) != len(o.Dependencies) {
		return false
	}
	for i := range s.Dependencies {
		if s.Dependencies[i] != o.Dependencies[i] {
			return false
		}
	}
	return s.Name == o.Name &&
		s.Namespace == o.Namespace &&
		s.Version == o.Version &&
		s.Annotation == o.Annotation &&
		s.Stateful == o.Stateful &&
		generics.Equal_StringSlice(s.Filters, o.Filters)
}

func (s PodSpec) Merge(o PodSpec) PodSpec {
	s.Containers = o.Containers
	s.Dependencies = o.Dependencies
	s.Filters = o.Filters
	s.Annotation = o.Annotation
	s.Stateful = o.Stateful
	s.Version += 1
	s.UpdatedAt = time.Now()
	s.PrevState = o.PrevState
	return s
}

func NewPodSpec(containerSpec ContainerSpec, otherSpecs ...ContainerSpec) PodSpec {
	cSpecs := make([]ContainerSpec, 1+len(otherSpecs))
	cSpecs[0] = containerSpec
	for i, cs := range otherSpecs {
		cSpecs[i+1] = cs
	}
	spec := PodSpec{
		Containers: cSpecs,
		PrevState:  NewPodPrevState(len(otherSpecs) + 1),
	}
	spec.Version = 1
	spec.CreatedAt = time.Now()
	spec.UpdatedAt = spec.CreatedAt
	return spec
}

type RestartPolicy int

const (
	RestartPolicyNever = iota
	RestartPolicyAlways
	RestartPolicyOnFail
)

func (rp RestartPolicy) String() string {
	switch rp {
	case RestartPolicyNever:
		return "RestartPolicyNever"
	case RestartPolicyAlways:
		return "RestartPolicyAlways"
	case RestartPolicyOnFail:
		return "RestartPolicyOnFail"
	default:
		return "Unknown RestartPolicy"
	}
}

type PodGroupPrevState struct {
	Nodes []string
	// we think a instance only have one ip, as now a instance only have one container.
	IPs []string
}

func (pgps PodGroupPrevState) Clone() PodGroupPrevState {
	newState := PodGroupPrevState{
		Nodes: make([]string, len(pgps.Nodes)),
		IPs:   make([]string, len(pgps.Nodes)),
	}
	copy(newState.Nodes, pgps.Nodes)
	copy(newState.IPs, pgps.IPs)
	return newState
}

func (pgps PodGroupPrevState) Reset(instanceNo int) PodGroupPrevState {
	newState := PodGroupPrevState{
		Nodes: make([]string, instanceNo),
		IPs:   make([]string, instanceNo),
	}
	copy(newState.Nodes, pgps.Nodes)
	copy(newState.IPs, pgps.IPs)
	return newState
}

func (pgps PodGroupPrevState) Length() int {
	if pgps.Nodes == nil {
		return 0
	}
	return len(pgps.Nodes)
}

type PodGroupSpec struct {
	ImSpec
	Pod           PodSpec
	NumInstances  int
	RestartPolicy RestartPolicy
}

func (spec PodGroupSpec) String() string {
	return fmt.Sprintf("PodGroup[name=%s, version=%d, #instances=%d, restart=%s]",
		spec.Name, spec.Version, spec.NumInstances, spec.RestartPolicy)
}

func (spec PodGroupSpec) Clone() PodGroupSpec {
	newSpec := spec
	newSpec.Pod = spec.Pod.Clone()
	return newSpec
}

func (spec PodGroupSpec) Equals(o PodGroupSpec) bool {
	return spec.Name == o.Name &&
		spec.Namespace == o.Namespace &&
		spec.Version == o.Version &&
		spec.Pod.Equals(o.Pod) &&
		spec.NumInstances == o.NumInstances &&
		spec.RestartPolicy == o.RestartPolicy
}

func (spec PodGroupSpec) VerifyParams() bool {
	verify := spec.Name != "" &&
		spec.Namespace != "" &&
		spec.NumInstances >= 0
	if !verify {
		return false
	}
	return spec.Pod.VerifyParams()
}

func NewPodGroupSpec(name string, namespace string, podSpec PodSpec, numInstances int) PodGroupSpec {
	spec := PodGroupSpec{
		Pod:          podSpec,
		NumInstances: numInstances,
	}
	spec.Name = name
	spec.Namespace = namespace
	spec.Version = 1
	spec.CreatedAt = time.Now()
	spec.UpdatedAt = spec.CreatedAt
	spec.Pod.ImSpec = spec.ImSpec
	return spec
}
