package api

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

// EventSource is a component which can hajick http connection and then send event to client.
type EventSource struct {
	conn        net.Conn
	closed      bool
	closeNotify chan bool
	locker      sync.Mutex
}

// EventData is a type which can be send by SendEvent(), SendEvent() will call Encode() and then write result into connection.
type EventData interface {
	Encode() ([]byte, error)
	Decode(content []byte) error
}

// NewEventSource create a new event source, event source will hijack w.conn; so you can not write anything to this responseWriter after calling this.
func NewEventSource(w http.ResponseWriter) (*EventSource, error) {
	var err error
	es := &EventSource{
		closeNotify: make(chan bool),
		closed:      false,
	}
	if hijacker, ok := w.(http.Hijacker); ok {
		es.conn, _, err = hijacker.Hijack()
	} else {
		return nil, fmt.Errorf("responseWriter do not support hijacker")
	}

	if _, err = es.conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n")); err != nil {
		es.Close()
		return nil, err
	}

	if _, err = es.conn.Write([]byte("Vary: Accept-Encoding\r\n")); err != nil {
		es.Close()
		return nil, err
	}

	if _, err = es.conn.Write([]byte("\r\n")); err != nil {
		es.Close()
		return nil, err
	}

	go func() { // this goroutine to monitor connection whether closed
		io.Copy(nil, es.conn)
		es.Close()
	}()
	return es, nil
}

// SendEvent send a event to client.
// data can be []byte, string or EventData. it should not contain '\n' in it, otherwize only the first line will be send.
// the given data can be string or []byte or EventData
func (es *EventSource) SendEvent(id uint64, event string, data interface{}) error {
	switch data.(type) {
	case string:
		es.conn.Write(generateMessage(id, event, data.(string)))
	case []byte:
		es.conn.Write(generateMessage(id, event, string(data.([]byte))))
	case EventData:
		content, err := data.(EventData).Encode()
		if err != nil {
			return err
		}
		es.conn.Write(generateMessage(id, event, string(content)))
	}
	return nil
}

// CloseNotify return a channel used to notify the connection's close event.
// the return channel will be closed when connection was closed, nothing send into it.
func (es *EventSource) CloseNotify() <-chan bool {
	return es.closeNotify
}

// Close the event source, this function also close the http connection.
func (es *EventSource) Close() {
	es.locker.Lock()
	defer es.locker.Unlock()
	if !es.closed {
		es.conn.Close()
		close(es.closeNotify)
		es.closed = true
	}
}

func generateMessage(id uint64, event, content string) []byte {
	var data bytes.Buffer
	data.WriteString(fmt.Sprintf("id: %d\n", id))
	if len(event) > 0 {
		data.WriteString(fmt.Sprintf("event: %s\n", strings.Replace(event, "\n", "", -1)))
	}
	if len(content) > 0 {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			data.WriteString(fmt.Sprintf("data: %s\n", line))
		}
	}
	data.WriteString("\n")
	return data.Bytes()
}
