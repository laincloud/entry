package engine

import (
	"fmt"
	"strconv"

	"strings"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/mijia/adoc"
	"github.com/mijia/go-generics"
	"github.com/mijia/sweb/log"
)

// podController is controlled by the podGroupController
type podController struct {
	spec PodSpec
	pod  Pod
}

func (pc *podController) String() string {
	return fmt.Sprintf("PodCtrl %s", pc.spec)
}

func (pc *podController) Deploy(cluster cluster.Cluster) {
	if pc.pod.State != RunStatePending {
		return
	}

	log.Infof("%s deploying", pc)
	start := time.Now()
	defer func() {
		pc.spec.Filters = []string{} // clear the filter
		pc.pod.UpdatedAt = time.Now()
		log.Infof("%s deployed, state=%+v, duration=%s", pc, pc.pod.ImRuntime, time.Now().Sub(start))
	}()

	pc.pod.Containers = make([]Container, len(pc.spec.Containers))
	pc.pod.LastError = ""
	filters := make([]string, 0, len(pc.spec.Filters)+1)
	filters = append(filters, pc.spec.Filters...)
	containerLabel := ContainerLabel{
		Name: pc.spec.Name,
	}
	filters = append(filters, containerLabel.NameAffnity())
	for i, cSpec := range pc.spec.Containers {
		log.Infof("%s create container, filter is %v", pc, filters)
		id, err := pc.createContainer(cluster, filters, i)
		if err != nil {
			log.Warnf("%s Cannot create container, error=%q, spec=%+v", pc, err, cSpec)
			pc.pod.State = RunStateFail
			pc.pod.LastError = fmt.Sprintf("Cannot create container, %s", err)
			return
		}
		if err := cluster.StartContainer(id); err != nil {
			log.Warnf("%s Cannot start container %s, %s", pc, id, err)
			pc.pod.State = RunStateFail
			pc.pod.LastError = fmt.Sprintf("Cannot start container, %s", err)
		}

		pc.pod.Containers[i].Id = id
		pc.refreshContainer(cluster, i)

		if i == 0 && pc.pod.Containers[0].NodeName != "" {
			filter := fmt.Sprintf("constraint:node==%s", pc.pod.Containers[0].NodeName)
			filters = append(filters, filter)
			pc.spec.PrevState.NodeName = pc.pod.Containers[i].NodeName
		}
		pc.spec.PrevState.IPs[i] = pc.pod.Containers[i].ContainerIp
	}
	if pc.pod.State == RunStatePending {
		pc.pod.State = RunStateSuccess
	}
}

func (pc *podController) Drift(cluster cluster.Cluster, fromNode, toNode string, force bool) bool {
	if pc.pod.State == RunStatePending {
		return false
	}
	if fromNode == toNode {
		return false
	}

	pod := pc.pod
	if len(pod.Containers) == 0 {
		return false
	}
	if pod.NodeName() != fromNode {
		return false
	}
	// if we have stateful pod, the containers cannot be drifted so far
	if pc.spec.IsStateful() && !force {
		log.Warnf("%s cannot be drifted since it is a stateful pod", pc)
		return false
	}

	log.Infof("%s drifting", pc)
	start := time.Now()
	defer func() {
		log.Infof("%s drifted, state=%+v, duration=%s", pc, pc.pod.ImRuntime, time.Now().Sub(start))
	}()

	pc.Remove(cluster)
	time.Sleep(10 * time.Second)
	pc.pod.State = RunStatePending
	pc.pod.DriftCount += 1
	if toNode == "" {
		pc.spec.Filters = append(pc.spec.Filters, fmt.Sprintf("constraint:node!=%s", fromNode))
	} else {
		pc.spec.Filters = append(pc.spec.Filters, fmt.Sprintf("constraint:node==%s", toNode))
	}
	pc.Deploy(cluster)
	return true
}

func (pc *podController) Remove(cluster cluster.Cluster) {
	log.Infof("%s removing", pc)
	start := time.Now()
	defer func() {
		log.Warnf("%s removed, duration=%s", pc, time.Now().Sub(start))
	}()

	pc.pod.LastError = ""
	for _, container := range pc.pod.Containers {
		if container.Id == "" {
			continue
		}
		// try to stop first, then remove it
		if err := cluster.StopContainer(container.Id, pc.spec.GetKillTimeout()); err != nil {
			log.Warnf("%s cannot stop the container %s, %s, remove it directly", pc, container.Id, err.Error())
		}
		if err := cluster.RemoveContainer(container.Id, true, false); err != nil {
			log.Warnf("%s Cannot remove the container %s, %s", pc, container.Id, err)
			pc.pod.LastError = fmt.Sprintf("Fail to remove container, %s", err)
		}
	}
	pc.pod.Containers = nil
	pc.pod.State = RunStateExit
	pc.pod.UpdatedAt = time.Now()
}

func (pc *podController) Stop(cluster cluster.Cluster) {
	if pc.pod.State != RunStateSuccess {
		return
	}
	log.Infof("%s stopping", pc)
	start := time.Now()
	defer func() {
		log.Infof("%s stopped, state=%+v, duration=%s", pc, pc.pod.ImRuntime, time.Now().Sub(start))
	}()

	pc.pod.LastError = ""

	for i, container := range pc.pod.Containers {
		if err := cluster.StopContainer(container.Id, pc.spec.GetKillTimeout()); err != nil {
			log.Warnf("%s Cannot stop the container %s, %s", pc, container.Id, err)
			pc.pod.State = RunStateFail
			pc.pod.LastError = fmt.Sprintf("Cannot stop container, %s", err)
		} else {
			pc.refreshContainer(cluster, i)
		}
	}
	pc.pod.UpdatedAt = time.Now()
}

func (pc *podController) Start(cluster cluster.Cluster) {
	if pc.pod.State == RunStatePending || pc.pod.State == RunStateSuccess {
		return
	}
	log.Infof("%s starting", pc)
	start := time.Now()
	defer func() {
		log.Infof("%s started, state=%+v, duration=%s", pc, pc.pod.ImRuntime, time.Now().Sub(start))
	}()
	pc.pod.State = RunStateSuccess
	pc.pod.LastError = ""
	for i, container := range pc.pod.Containers {
		if err := cluster.StartContainer(container.Id); err != nil {
			log.Warnf("%s Cannot start the container %s, %s", pc, container.Id, err)
			pc.pod.State = RunStateFail
			pc.pod.LastError = fmt.Sprintf("Cannot start container, %s", err)
		} else {
			pc.refreshContainer(cluster, i)
		}
	}
	pc.pod.UpdatedAt = time.Now()
}

func (pc *podController) Refresh(cluster cluster.Cluster) {
	log.Infof("%s refreshing", pc)
	start := time.Now()
	defer func() {
		log.Infof("%s refreshed, state=%+v, duration=%s", pc, pc.pod.ImRuntime, time.Now().Sub(start))
	}()

	pc.pod.State = RunStateSuccess
	pc.pod.LastError = ""

	for i := 0; i < len(pc.spec.Containers); i += 1 {
		pc.refreshContainer(cluster, i)
	}
	pc.pod.UpdatedAt = time.Now()
}

// tryCorrectIPAddress try to correct container's ip address to given ip. return true if successed, otherwise return false.
func (pc *podController) tryCorrectIPAddress(c cluster.Cluster, id, fromIP, toIP string) bool {
	if err := c.DisconnectContainer(pc.spec.Namespace, id, true); err != nil {
		log.Errorf("%s fail to disconnect network %s to container %s, %s", pc, pc.spec.Namespace, id, err.Error())
		// do not return false, try to connect.
	}
	if err := c.ConnectContainer(pc.spec.Namespace, id, toIP); err != nil {
		log.Errorf("%s fail to connect network %s to container %s by using IP %s, %s", pc, pc.spec.Namespace, id, toIP, err.Error())
		log.Infof("%s try to recover network using old ip %s", pc, fromIP)
		if err := c.ConnectContainer(pc.spec.Namespace, id, fromIP); err != nil {
			log.Errorf("%s fail to recover network %s to container %s by using oldIP %s, %s, now container ip lost, give up!", pc, pc.spec.Namespace, id, fromIP, err.Error())
			log.Warnf("%s can not set any ip for container %s, give ip!!!")
		}
		return false
	}
	return true
}

func (pc *podController) refreshContainer(kluster cluster.Cluster, index int) {
	if index < 0 || index >= len(pc.pod.Containers) {
		return
	}
	id := pc.pod.Containers[index].Id
	if id == "" {
		pc.pod.State = RunStateMissing
		pc.pod.LastError = fmt.Sprintf("Missing container, without the container id.")
		return
	}

	spec := pc.spec.Containers[index]
	if info, err := kluster.InspectContainer(id); err != nil {
		if adoc.IsNotFound(err) {
			log.Warnf("%s We found some missing container %s, %s", pc, id, err)
			pc.pod.State = RunStateMissing
			pc.pod.LastError = fmt.Sprintf("Missing container %q, %s", id, err)
		} else {
			log.Warnf("%s Failed to inspect container %s, %s", pc, id, err)
			pc.pod.State = RunStateFail
			pc.pod.LastError = fmt.Sprintf("Cannot inspect the container, %s", err)
		}
	} else {
		prevIP, nowIP := pc.spec.PrevState.IPs[index], info.NetworkSettings.Networks[pc.spec.Namespace].IPAddress

		// NOTE: if the container's ip is not equal to prev ip, try to correct it; if failed, accpet new ip
		if prevIP != "" && prevIP != nowIP {
			log.Warnf("%s find the IP changed, prev is %s, but now is %s, try to correct it", pc, prevIP, nowIP)
			if !pc.tryCorrectIPAddress(kluster, id, nowIP, prevIP) {
				log.Warnf("%s fail to correct container ip to %s, accpet new ip %s.", pc, prevIP, nowIP)
			} else {
				nowIP = prevIP
			}
		}

		container := Container{
			Id:            id,
			Runtime:       info,
			NodeName:      info.Node.Name,
			NodeIp:        info.Node.IP,
			Protocol:      "tcp",
			ContainerIp:   nowIP,
			ContainerPort: spec.Expose,
		}
		// FIXME: until we start working on the multiple ports
		if ports, ok := info.NetworkSettings.Ports[fmt.Sprintf("%d/tcp", spec.Expose)]; ok && len(ports) > 0 {
			if port, err := strconv.Atoi(ports[0].HostPort); err == nil {
				container.NodePort = port
			}
		}

		pc.spec.PrevState.NodeName = info.Node.Name
		pc.spec.PrevState.IPs[index] = container.ContainerIp
		pc.pod.Containers[index] = container
		state := info.State
		if !state.Running {
			if state.ExitCode == 0 {
				pc.pod.State = RunStateExit
			} else {
				pc.pod.State = RunStateFail
				pc.pod.LastError = state.Error
			}
		}
	}
}

func (pc *podController) createContainer(cluster cluster.Cluster, filters []string, index int) (string, error) {
	cc := pc.createContainerConfig(filters, index)
	hc := pc.createHostConfig(index)
	nc := pc.createNetworkingConfig(index)
	name := pc.createContainerName(index)

	// FIXME: do we still need to rename this?
	//if container, err := cluster.InspectContainer(name); err == nil {
	//We got a same-named container which we need to rename it
	//newName := fmt.Sprintf("%s-lain_did_it_%d", name, time.Now().Unix())
	//log.Warnf("%s we found container has the same name %s, try to rename it to %s", pc, name, newName)
	//if err := cluster.RenameContainer(container.Id, newName); err != nil {
	//log.Warnf("%s Failed to rename the container as we need it, %s", pc, err)
	//}
	//}
	return cluster.CreateContainer(cc, hc, nc, name)
}

func (pc *podController) createContainerConfig(filters []string, index int) adoc.ContainerConfig {
	podSpec := pc.spec
	spec := podSpec.Containers[index]

	volumes := make(map[string]struct{})
	for _, v := range spec.Volumes {
		volumes[v] = struct{}{}
	}
	for _, sv := range spec.SystemVolumes {
		parts := strings.Split(sv, ":")
		if len(parts) > 1 {
			volumes[parts[1]] = struct{}{}
		}
	}

	injectEnvs := append(spec.Env, []string{
		fmt.Sprintf("DEPLOYD_POD_INSTANCE_NO=%d", pc.pod.InstanceNo),
		fmt.Sprintf("DEPLOYD_POD_NAME=%s", pc.spec.Name),
		fmt.Sprintf("DEPLOYD_POD_NAMESPACE=%s", pc.spec.Namespace),
	}...)
	injectEnvs = append(injectEnvs, filters...)

	containerLabel := ContainerLabel{
		Name:           podSpec.Name,
		Namespace:      podSpec.Namespace,
		InstanceNo:     pc.pod.InstanceNo,
		Version:        podSpec.Version,
		DriftCount:     pc.pod.DriftCount,
		ContainerIndex: index,
		Annotation:     podSpec.Annotation,
	}

	cc := adoc.ContainerConfig{
		Image:      spec.Image,
		Cmd:        spec.Command,
		Env:        injectEnvs,
		Memory:     spec.MemoryLimit,
		MemorySwap: spec.MemoryLimit,
		CpuShares:  spec.CpuLimit,
		User:       spec.User,
		WorkingDir: spec.WorkingDir,
		Volumes:    volumes,
		Entrypoint: spec.Entrypoint,
		Labels:     containerLabel.Label2Maps(),
	}
	if spec.Expose > 0 {
		cc.ExposedPorts = map[string]struct{}{
			fmt.Sprintf("%d/tcp", spec.Expose): struct{}{},
		}
	}
	return cc
}

func (pc *podController) createHostConfig(index int) adoc.HostConfig {
	podSpec := pc.spec
	spec := podSpec.Containers[index]
	hc := adoc.HostConfig{}
	if spec.Expose > 0 {
		hc.PortBindings = map[string][]adoc.PortBinding{
			fmt.Sprintf("%d/tcp", spec.Expose): []adoc.PortBinding{
				adoc.PortBinding{},
			},
		}
	}
	if len(spec.Volumes) > 0 {
		binds := make([]string, len(spec.Volumes))
		for i, v := range spec.Volumes {
			// /data/lain/volumes/hello/hello.proc.web.foo/1/{c0}/{v:v}
			if len(podSpec.Containers) > 1 {
				binds[i] = fmt.Sprintf("%s/%s/%s/%d/c%d/%s:%s", kLainVolumeRoot, podSpec.Namespace, podSpec.Name, pc.pod.InstanceNo, index, v, v)
			} else {
				binds[i] = fmt.Sprintf("%s/%s/%s/%d/%s:%s", kLainVolumeRoot, podSpec.Namespace, podSpec.Name, pc.pod.InstanceNo, v, v)
			}
		}
		hc.Binds = binds
	}
	hc.Binds = append(hc.Binds, spec.SystemVolumes...)
	hc.NetworkMode = podSpec.Namespace
	if len(spec.DnsSearch) > 0 {
		hc.DnsSearch = generics.Clone_StringSlice(spec.DnsSearch)
	}
	if spec.LogConfig.Type != "" {
		hc.LogConfig.Type = spec.LogConfig.Type
		hc.LogConfig.Config = generics.Clone_StringStringMap(spec.LogConfig.Config)
	}
	return hc
}

func (pc *podController) createContainerName(index int) string {
	segs := make([]string, 0, 2)
	segs = append(segs, fmt.Sprintf("%s.v%d-i%d-d%d", pc.spec.Name, pc.spec.Version, pc.pod.InstanceNo, pc.pod.DriftCount))
	if len(pc.spec.Containers) > 1 {
		segs = append(segs, fmt.Sprintf("c%d", index))
	}
	return strings.Join(segs, "-")
}

func (pc *podController) createNetworkingConfig(index int) adoc.NetworkingConfig {
	podSpec := pc.spec
	net := podSpec.Namespace
	nc := adoc.NetworkingConfig{}
	ipamc := adoc.IPAMConfig{}
	ipamc.IPv4Address = pc.spec.PrevState.IPs[index]
	nc.EndpointsConfig = map[string]adoc.EndpointConfig{
		net: adoc.EndpointConfig{
			ipamc,
		},
	}
	return nc
}
