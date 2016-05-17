package adoc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (client *DockerClient) StopMonitor(monitorId int64) {
	client.monitorLock.Lock()
	defer client.monitorLock.Unlock()
	delete(client.monitors, monitorId)
}

func (client *DockerClient) newMonitorItem() int64 {
	client.monitorLock.Lock()
	defer client.monitorLock.Unlock()

	var monitorId int64
	for trial := 5; trial > 0; trial -= 1 {
		monitorId = random.Int63()
		if _, ok := client.monitors[monitorId]; !ok {
			client.monitors[monitorId] = struct{}{}
			break
		}
	}
	// we have some change to conflict, but I think maybe we live with that
	return monitorId
}

type EventCallback func(event Event, err error)

func (client *DockerClient) MonitorEvents(filters string, callback EventCallback) int64 {
	v := url.Values{}
	if filters != "" {
		v.Set("filters", filters)
	}
	uri := "events"
	if len(v) > 0 {
		uri += "?" + v.Encode()
	}
	monitorId := client.newMonitorItem()
	go client.monitorEvents(monitorId, uri, callback)
	return monitorId
}

// will be running inside a goroutine
func (client *DockerClient) monitorEvents(monitorId int64, uri string, callback EventCallback) {
	err := client.sendRequestCallback("GET", uri, nil, nil, func(resp *http.Response) error {
		decoder := json.NewDecoder(resp.Body)
		client.monitorLock.RLock()
		_, toContinue := client.monitors[monitorId]
		client.monitorLock.RUnlock()
		for toContinue {
			var event Event
			if err := decoder.Decode(&event); err != nil {
				return err
			}
			callback(event, nil)
			client.monitorLock.RLock()
			_, toContinue = client.monitors[monitorId]
			client.monitorLock.RUnlock()
		}
		return nil
	}, true)
	if err != nil && err != io.EOF {
		callback(Event{}, err)
	}
}

type StatsCallback func(stats Stats, err error)

func (client *DockerClient) MonitorStats(containerId string, callback StatsCallback) int64 {
	uri := fmt.Sprintf("containers/%s/stats", containerId)
	monitorId := client.newMonitorItem()
	go client.monitorStats(monitorId, uri, callback)
	return monitorId
}

func (client *DockerClient) monitorStats(monitorId int64, uri string, callback StatsCallback) {
	err := client.sendRequestCallback("GET", uri, nil, nil, func(resp *http.Response) error {
		decoder := json.NewDecoder(resp.Body)
		client.monitorLock.RLock()
		_, toContinue := client.monitors[monitorId]
		client.monitorLock.RUnlock()
		for toContinue {
			var stats Stats
			if err := decoder.Decode(&stats); err != nil {
				return err
			}
			callback(stats, nil)
			client.monitorLock.RLock()
			_, toContinue = client.monitors[monitorId]
			client.monitorLock.RUnlock()
		}
		return nil
	}, true)
	if err != nil && err != io.EOF {
		callback(Stats{}, err)
	}
}
