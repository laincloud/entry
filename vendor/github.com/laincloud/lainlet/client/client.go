package client

import (
	"bufio"
	"bytes"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"github.com/laincloud/lainlet/api/v2"
	"github.com/laincloud/lainlet/watcher/container"
	"github.com/laincloud/lainlet/watcher/nodes"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// all the kinds of event
const (
	ERROR     string = "error"
	UPDATE           = "update"
	DELETE           = "delete"
	INIT             = "init"
	HEARTBEAT        = "heartbeat"
)

// The Data type return by /v2/configwatcher
type Config map[string]string

// The Data type return by /v2/containers
type Containers map[string]ContainerInfo
type ContainerInfo container.Info

// The Data type return by /v2/coreinfowatcher
type CoreInfos map[string]CoreInfo
type CoreInfo v2.CoreInfo

// The Data type return by /v2/depends
type Depends map[string]map[string]map[string]DependsItem
type DependsItem v2.DependsItem

// The Data type return by /v2/localspecquery
type LocalSpecs []string

// The Data type return by /v2/nodes
type NodesInfo map[string]NodeInfo
type NodeInfo nodes.NodeInfo

// The Data type return by /v2/podgroupwatcher
type PodGroups []PodGroup
type PodGroup v2.PodGroup

// The Data type return by /v2/proxywatcher
type Proxy map[string]ProcInfo
type ProcInfo v2.ProcInfo

// The Data type return by /v2/configwatcher?target=vips
type JSONVirtualIpPortConfig struct {
	App      string `json:"app"`
	Proc     string `json:"proc"`
	Port     string `json:"port"`
	Proto    string `json:"proto"`
	ProcType string `json:"proctype"`
}

// the response returned by watch action
type Response struct {
	Id    int64  // The event Id, start from 1, and incresed by 1 when event is UPDATE or DELETE, it's always 0 when event is ERROR or HEARTBEAT
	Event string // the event name

	// the returned data return by watch request, the Data is a json-format.
	// you can import `github.com/laincloud/api/v2`, and using the coresponding type to Decode.
	Data     []byte
	finished bool // check if one response was completed, because it's field was set by the data read from response body line by line
}

type Client struct {
	addr string
}

// create a new client, addr is lainlet address such as "127.0.0.1:9001"
func New(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

// get data from the given uri, return the []byte read from http response body.
// It return error if fail to send http request to lainlet
func (c *Client) Get(uri string, timeout time.Duration) ([]byte, error) {
	if timeout < 0 {
		timeout = 0
	}
	reader, err := c.Do(uri, timeout, false)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

// watch request to the given uri.
// The return channel will be closed when the context was canceled or the http connection was closed.
// it return error when fail to send request to lainlet.
func (c *Client) Watch(uri string, ctx context.Context) (<-chan *Response, error) {
	reader, err := c.Do(uri, 0, true)
	if err != nil {
		return nil, err
	}
	respCh := make(chan *Response)
	stop := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-stop:
		}
		reader.Close()
	}()
	go func() {
		defer close(stop)
		defer close(respCh)
		var (
			newone bool = true // if need to create a new response
			resp   *Response
		)
		buf := bufio.NewReader(reader)
		for {
			if newone {
				resp = new(Response)
				resp.finished = false
				newone = false
			}
			line, err := buf.ReadBytes('\n')
			if err != nil {
				if err != io.EOF || len(line) == 0 {
					// got a unexpected(not io.EOF) error, or read none data with io.EOF
					return
				}
			}
			if pureLine := bytes.TrimSpace(line); len(pureLine) == 0 {
				continue // empty line
			} else {
				fields := bytes.SplitN(pureLine, []byte{':'}, 2)
				if len(fields) < 2 {
					continue
				}
				switch key := string(bytes.TrimSpace(fields[0])); key {
				case "id":
					if id, err := strconv.ParseInt(string(bytes.TrimSpace(fields[1])), 10, 64); err == nil {
						resp.Id = id
					}
				case "event":
					resp.Event = string(bytes.TrimSpace(fields[1]))
					if resp.Event == "heartbeat" {
						resp.finished = true
					}
				case "data":
					resp.Data = bytes.TrimSpace(fields[1])
					resp.finished = true
				}
			}
			if resp.finished {
				respCh <- resp
				newone = true
			}
		}
	}()
	return (<-chan *Response)(respCh), nil
}

// send a http request
func (c *Client) Do(uri string, timeout time.Duration, watch bool) (io.ReadCloser, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	u.Scheme = "http"
	u.Host = c.addr

	q := u.Query()
	if watch {
		q.Set("watch", "1")
	} else {
		q.Set("watch", "0")
	}
	u.RawQuery = q.Encode()

	resp, err := (&http.Client{Timeout: timeout}).Get(u.String())
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
