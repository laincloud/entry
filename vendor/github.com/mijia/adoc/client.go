package adoc

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type AuthConfig struct {
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

func (auth AuthConfig) Encode() string {
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(auth)
	return base64.URLEncoding.EncodeToString(buffer.Bytes())
}

type Error struct {
	StatusCode int
	Status     string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Status)
}

func IsNotFound(err error) bool {
	if adocErr, ok := err.(Error); ok {
		return adocErr.StatusCode == 404
	}
	return false
}

func IsServerInternalError(err error) bool {
	if adocErr, ok := err.(Error); ok {
		return adocErr.StatusCode == 500
	}
	return false
}

const (
	kDefaultApiVersion = "v1.17"
	kDefaultTimeout    = 30
	kDefaultRWTimeout  = 60
)

var apiVersions = map[string]bool{
	"v1.17": true,
	"v1.18": true,
}

type DockerClient struct {
	daemonUrl      *url.URL
	httpClient     *http.Client
	longpollClient *http.Client
	tlsConfig      *tls.Config
	apiVersion     string
	isSwarm        bool

	monitorLock sync.RWMutex
	monitors    map[int64]struct{}
}

func NewSwarmClient(swarmUrl string, tlsConfig *tls.Config, apiVersion ...string) (*DockerClient, error) {
	return NewSwarmClientTimeout(swarmUrl, tlsConfig,
		time.Duration(kDefaultTimeout*time.Second),
		time.Duration(kDefaultRWTimeout*time.Second),
		apiVersion...)
}

func NewSwarmClientTimeout(swarmUrl string, tlsConfig *tls.Config, timeout time.Duration, rwTimeout time.Duration, apiVersion ...string) (*DockerClient, error) {
	docker, err := NewDockerClientTimeout(swarmUrl, tlsConfig, timeout, rwTimeout, apiVersion...)
	docker.isSwarm = true
	return docker, err
}

func NewDockerClient(daemonUrl string, tlsConfig *tls.Config, apiVersion ...string) (*DockerClient, error) {
	return NewDockerClientTimeout(daemonUrl, tlsConfig,
		time.Duration(kDefaultTimeout*time.Second),
		time.Duration(kDefaultRWTimeout*time.Second),
		apiVersion...)
}

func NewDockerClientTimeout(daemonUrl string, tlsConfig *tls.Config, timeout time.Duration, rwTimeout time.Duration, apiVersion ...string) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Scheme == "tcp" {
		if tlsConfig == nil {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}
	httpClient := newHttpClient(u, tlsConfig, timeout, rwTimeout)
	longpollClient := newHttpClient(u, tlsConfig, timeout, 0)
	clientApiVersion := kDefaultApiVersion
	if len(apiVersion) > 0 && apiVersion[0] != "" {
		clientApiVersion = apiVersion[0]
		if !strings.HasPrefix(clientApiVersion, "v") {
			clientApiVersion = "v" + clientApiVersion
		}
	}
	if _, checked := apiVersions[clientApiVersion]; !checked {
		err = fmt.Errorf("*WARNING: Adoc haven't check out if the remote api version %s is supported, maybe not stable, but you can keep using the client anyway.", clientApiVersion)
	}
	return &DockerClient{
		daemonUrl:      u,
		httpClient:     httpClient,
		longpollClient: longpollClient,
		tlsConfig:      tlsConfig,
		apiVersion:     clientApiVersion,
		monitors:       make(map[int64]struct{}),
	}, err
}

type responseCallback func(resp *http.Response) error

func (client *DockerClient) sendRequestCallback(method string, path string, body []byte, headers map[string]string, callback responseCallback, isLongpoll ...bool) error {
	b := bytes.NewBuffer(body)
	urlPath := fmt.Sprintf("%s/%s/%s", client.daemonUrl.String(), client.apiVersion, path)
	logger.Debugf("SendRequest %q, [%s]", method, urlPath)
	req, err := http.NewRequest(method, urlPath, b)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}
	httpClient := client.httpClient
	if len(isLongpoll) > 0 && isLongpoll[0] {
		httpClient = client.longpollClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "connection refused") && client.tlsConfig == nil {
			return fmt.Errorf("%v. Are you trying to connect to a TLS-enabled daemon without TLS?", err)
		}
		return err
	}
	if resp.StatusCode >= 400 {
		return Error{resp.StatusCode, resp.Status}
	}

	defer resp.Body.Close()
	return callback(resp)
}

func (client *DockerClient) sendRequest(method string, path string, body []byte, headers map[string]string, isLongpoll ...bool) ([]byte, error) {
	var data []byte
	err := client.sendRequestCallback(method, path, body, headers, func(resp *http.Response) error {
		var cbErr error
		if data, cbErr = ioutil.ReadAll(resp.Body); cbErr != nil {
			return cbErr
		}
		return nil
	}, isLongpoll...)
	return data, err
}

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}
