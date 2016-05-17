# Another DOcker Client in Go

ADOC is Another DOcker Client implemented in Go.

Supports:

1. Docker 1.5/1.6 with remote api versions of 1.17, 1.18
1. Add supports for Swarm, only minor differences with Docker API

Example:

```
	docker, err := adoc.NewDockerClient("tcp://<docker_tcp_port>", nil)

	// or you can pass the api version that you want to use
	docker, err := adoc.NewDockerClient("tcp://<docker_tcp_port>", nil, "1.18")
	
	// or you can use the swarm client
	docker, err := adoc.NewSwarmClient("tcp://<swarm_tcp_port>", nil)
	
	version, err := docker.Version()
	info, err := docker.Info()
	
	// create a container and start it
	containerConf := adoc.ContainerConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"python", "app.py"},
		Image:        "training/webapp",
	}
	hostConf := adoc.HostConfig{
		PortBindings: map[string][]PortBinding{
			"5000/tcp": []PortBinding{
				PortBinding{},
			},
		},
	}
	id, err := docker.CreateContainer(containerConf, hostConf)
	err := docker.StartContainer(id)

	// Pull, inspect and remove an Image
	err := docker.PullImage("busybox", "latest")
	image, err := docker.InspectImage("busybox")
	err := docker.RemoveImage("busybox", false, false)
	
	// Monitor some stats from a running container
	monitorId := docker.MonitorStats(containerId, func(stats adoc.Stats, err error) {
		if err == nil {
			fmt.Println("Stats CpuUsage.PercpuUsage", stats.CpuStats.CpuUsage.PercpuUsage)
		}
	})
	time.Sleep(30 * time.Second)
	docker.StopMonitor(monitorId)
	
	// Polling some events happend since an hour agao until 5 minutes ago
	events, err := docker.EventsSince("", time.Hour, 5*time.Minute)
	
	// Or monitor all the events
	monitorId := docker.MonitorEvents("", func(event adoc.Event, err error) {
		if err == nil {
			fmt.Println("Event", event)
		}
	})
	time.Sleep(1 * time.Minute)
	docker.StopMonitor(monitorId)
	
	...
	
```