package adoc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// This part contains the misc apis listed in
// https://docs.docker.com/reference/api/docker_remote_api_v1.17/#23-misc

type Version struct {
	ApiVersion    string
	GitCommit     string
	GoVersion     string
	Version       string
	Os            string // v1.18
	Arch          string // v1.18
	KernelVersion string // v1.18
}

type SwarmNodeInfo struct {
	Name       string
	Address    string
	Containers int64
	CPUs       int
	UsedCPUs   int
	Memory     int64
	UsedMemory int64
}

type SwarmInfo struct {
	Role       string
	Containers int64
	Strategy   string
	Filters    string
	Nodes      []SwarmNodeInfo
}

type DockerInfo struct {
	Containers      int64
	DockerRootDir   string
	Driver          string
	DriverStatus    [][2]string
	ExecutionDriver string
	ID              string
	//IPv4Forwarding     int
	Images             int64
	IndexServerAddress string
	InitPath           string
	InitSha1           string
	KernelVersion      string
	Labels             []string
	MemTotal           int64
	//MemoryLimit        int
	NCPU            int64
	NEventsListener int64
	NFd             int64
	NGoroutines     int64
	Name            string
	OperatingSystem string
	//SwapLimit          int
	HttpProxy  string    // v1.18
	HttpsProxy string    // v1.18
	NoProxy    string    // v1.18
	SystemTime time.Time // v1.18
	//Debug              bool // this will conflict with docker api and swarm api, fuck
}

type ExecConfig struct {
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	Tty          bool
	Cmd          []string
}

func (client *DockerClient) Version() (Version, error) {
	var ret Version
	if data, err := client.sendRequest("GET", "version", nil, nil); err != nil {
		return Version{}, err
	} else {
		err := json.Unmarshal(data, &ret)
		return ret, err
	}
}

func (client *DockerClient) IsSwarm() bool {
	return client.isSwarm
}

func (client *DockerClient) SwarmInfo() (SwarmInfo, error) {
	var ret SwarmInfo
	if !client.isSwarm {
		return ret, fmt.Errorf("The client is not a swarm client, please use Info()")
	}
	info, err := client.Info()
	if err != nil {
		return ret, err
	}
	ret.Containers = info.Containers
	offset := 0
	for ; offset < len(info.DriverStatus); offset += 1 {
		key, value := info.DriverStatus[offset][0], info.DriverStatus[offset][1]
		if strings.HasSuffix(key, "Role") {
			ret.Role = value
		}
		if strings.HasSuffix(key, "Strategy") {
			ret.Strategy = value
		}
		if strings.HasSuffix(key, "Filters") {
			ret.Filters = value
		}
		if strings.HasSuffix(key, "Nodes") {
			break
		}
	}

	nodeCount, _ := strconv.Atoi(info.DriverStatus[offset][1])
	ret.Nodes = make([]SwarmNodeInfo, nodeCount)
	offset += 1
	for i := 0; i < nodeCount; i += 1 {
		if nodeInfo, err := parseSwarmNodeInfo(info.DriverStatus[offset : offset+5]); err == nil {
			ret.Nodes[i] = nodeInfo
		}
		offset += 5
	}
	return ret, nil
}

func parseSwarmNodeInfo(data [][2]string) (ret SwarmNodeInfo, parseErr error) {
	defer func() {
		if err := recover(); err != nil {
			parseErr = fmt.Errorf("Paniced when parse swarm node info, the protocol maybe changed, %s", err)
			logger.Warnf(parseErr.Error())
		}
	}()
	ret.Name = data[0][0]
	ret.Address = data[0][1]
	ret.Containers, _ = strconv.ParseInt(data[1][1], 10, 64)

	cpuInfo := strings.Split(data[2][1], "/")
	ret.UsedCPUs, _ = strconv.Atoi(strings.TrimSpace(cpuInfo[0]))
	ret.CPUs, _ = strconv.Atoi(strings.TrimSpace(cpuInfo[1]))

	memInfo := strings.Split(data[3][1], "/")
	ret.UsedMemory, _ = ParseBytesSize(memInfo[0])
	ret.Memory, _ = ParseBytesSize(memInfo[1])
	return
}

func (client *DockerClient) Info() (DockerInfo, error) {
	var ret DockerInfo
	if data, err := client.sendRequest("GET", "info", nil, nil); err != nil {
		return ret, err
	} else {
		err := json.Unmarshal(data, &ret)
		return ret, err
	}
}

func (client *DockerClient) Ping() (bool, error) {
	if data, err := client.sendRequest("GET", "_ping", nil, nil); err != nil {
		return false, err
	} else {
		return string(data) == "OK", nil
	}
}

func (client *DockerClient) CreateExec(id string, execConfig ExecConfig) (string, error) {
	if body, err := json.Marshal(execConfig); err != nil {
		return "", err
	} else {
		uri := fmt.Sprintf("containers/%s/exec", id)
		if data, err := client.sendRequest("POST", uri, body, nil); err != nil {
			return "", err
		} else {
			var ret map[string]interface{}
			if err := json.Unmarshal(data, &ret); err != nil {
				return "", err
			}
			if execId, ok := ret["Id"]; ok {
				return execId.(string), nil
			}
			return "", fmt.Errorf("Cannot find Id field inside result object, %+v", ret)
		}
	}
}

func (client *DockerClient) StartExec(execId string, detach, tty bool) ([]byte, error) {
	params := map[string]bool{
		"Detach": detach,
		"Tty":    tty,
	}
	if body, err := json.Marshal(params); err != nil {
		return nil, err
	} else {
		uri := fmt.Sprintf("exec/%s/start", execId)
		return client.sendRequest("POST", uri, body, nil)
	}
}

// Missing apis for
// auth
// commit: Create a new image from a container's changes
// events: Monitor Docker's events
// images/(name)/get: Get a tarball containing all images in a repository
// images/get: Get a tarball containing all images.
// images/load: Load a tarball with a set of images and tags into docker
// exec/(id)/resize
// exec/(id)/json
