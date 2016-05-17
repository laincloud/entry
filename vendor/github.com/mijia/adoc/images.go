package adoc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// This part contains apis for the images listed in
// https://docs.docker.com/reference/api/docker_remote_api_v1.17/#22-images

type Image struct {
	Created     int64
	Id          string
	Labels      map[string]string
	ParentId    string
	RepoTags    []string
	Size        int64
	VirtualSize int64
	RepoDigests []string // v1.18
}

type ImageDetail struct {
	Architecture    string
	Author          string
	Comment         string
	Container       string
	ContainerConfig ContainerConfig
	Created         time.Time
	DockerVersion   string
	Id              string
	Os              string
	Parent          string
	Size            int64
	VirtualSize     int64
	//Config          ContainerConfig // don't know what this is for
}

func (client *DockerClient) ListImages(showAll bool, filters ...string) ([]Image, error) {
	v := url.Values{}
	v.Set("all", formatBoolToIntString(showAll))
	if len(filters) > 0 && filters[0] != "" {
		v.Set("filters", filters[0])
	}
	uri := fmt.Sprintf("images/json?%s", v.Encode())
	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return nil, err
	} else {
		var images []Image
		err := json.Unmarshal(data, &images)
		return images, err
	}
}

func (client *DockerClient) InspectImage(name string) (ImageDetail, error) {
	var ret ImageDetail
	uri := fmt.Sprintf("images/%s/json", name)
	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return ret, err
	} else {
		err := json.Unmarshal(data, &ret)
		return ret, err
	}
}

func (client *DockerClient) PullImage(name string, tag string, authConfig ...AuthConfig) error {
	v := url.Values{}
	v.Set("fromImage", name)
	v.Set("tag", tag)
	uri := fmt.Sprintf("images/create?%s", v.Encode())
	header := make(map[string]string)
	if len(authConfig) > 0 {
		header["X-Registry-Auth"] = authConfig[0].Encode()
	}
	err := client.sendRequestCallback("POST", uri, nil, header, func(resp *http.Response) error {
		var status map[string]interface{}
		var cbErr error
		decoder := json.NewDecoder(resp.Body)
		for ; cbErr == nil; cbErr = decoder.Decode(&status) {
		}
		if cbErr != io.EOF {
			return cbErr
		}
		if errMsg, ok := status["error"]; ok {
			return fmt.Errorf("Pull image error: %s", errMsg)
		}
		return nil
	}, true)
	return err
}

func (client *DockerClient) RemoveImage(name string, force, noprune bool) error {
	v := url.Values{}
	v.Set("force", formatBoolToIntString(force))
	v.Set("noprune", formatBoolToIntString(noprune))
	uri := fmt.Sprintf("images/%s?%s", name, v.Encode())
	_, err := client.sendRequest("DELETE", uri, nil, nil)
	return err
}

// Missing apis for
// build: Build image from a Dockerfile
// images/(name)/history
// images/(name)/push
// images/(name)/tag
// images/search
