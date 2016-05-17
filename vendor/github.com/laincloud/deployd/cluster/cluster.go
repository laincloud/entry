package cluster

import "github.com/mijia/adoc"

type Node struct {
	Name       string
	Address    string
	Containers int64
	CPUs       int
	UsedCPUs   int
	Memory     int64
	UsedMemory int64
}

func (n Node) SpareCPUs() int {
	return n.CPUs - n.UsedCPUs
}

func (n Node) SpareMemory() int64 {
	return n.Memory - n.UsedMemory
}

type Cluster interface {
	GetResources() ([]Node, error)

	ListContainers(showAll bool, showSize bool, filters ...string) ([]adoc.Container, error)
	CreateContainer(cc adoc.ContainerConfig, hc adoc.HostConfig, nc adoc.NetworkingConfig, name ...string) (string, error)
	ConnectContainer(networkName string, id string, ipAddr string) error
	DisconnectContainer(networkName string, id string, force bool) error
	StartContainer(id string) error
	StopContainer(id string, timeout ...int) error
	InspectContainer(id string) (adoc.ContainerDetail, error)
	RemoveContainer(id string, force bool, volumes bool) error
	RenameContainer(id string, name string) error

	MonitorEvents(filter string, callback adoc.EventCallback) int64
	StopMonitor(monitorId int64)
}
