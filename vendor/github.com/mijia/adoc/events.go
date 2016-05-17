package adoc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DockerEventCreate      = "create"
	DockerEventDestroy     = "destroy"
	DockerEventDie         = "die"
	DockerEventExecCreate  = "exec_create"
	DockerEventExecStart   = "exec_start"
	DockerEventExport      = "export"
	DockerEventKill        = "kill"
	DockerEventOOM         = "oom"
	DockerEventPause       = "pause"
	DockerEventRestart     = "restart"
	DockerEventStart       = "start"
	DockerEventStop        = "stop"
	DockerEventUnpause     = "unpause"
	DockerEventRename      = "rename"
	DockerEventImageUntag  = "untag"
	DockerEventImageDelete = "delete"
)

type Event struct {
	Id     string    `json:"id"`
	Status string    `json:"status"`
	From   string    `json:"from"`
	Time   int64     `json:"time"`
	Node   SwarmNode `json:"node,omitempty"`
}

func (client *DockerClient) EventsSince(filters string, since time.Duration, until ...time.Duration) ([]Event, error) {
	if client.isSwarm {
		return nil, fmt.Errorf("Swarm doesn't support the events polling mode.")
	}

	v := url.Values{}
	if filters != "" {
		v.Set("filters", filters)
	}
	now := time.Now()
	v.Set("since", fmt.Sprintf("%d", tsFromNow(now, since)))
	if len(until) > 0 {
		v.Set("until", fmt.Sprintf("%d", tsFromNow(now, until[0])))
	}
	uri := fmt.Sprintf("events?%s", v.Encode())

	events := make([]Event, 0)
	err := client.sendRequestCallback("GET", uri, nil, nil, func(resp *http.Response) error {
		var event Event
		var cbErr error
		decoder := json.NewDecoder(resp.Body)
		for ; cbErr == nil; cbErr = decoder.Decode(&event) {
			// Not sure about why there will be an empty event first
			if cbErr == nil && event.Status != "" {
				events = append(events, event)
			}
		}
		if cbErr != io.EOF {
			return cbErr
		}
		return nil
	})
	return events, err
}

func tsFromNow(now time.Time, duration time.Duration) int64 {
	return now.Add(-1 * duration).Unix()
}
