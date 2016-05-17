package adoc

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrInvalidHeader = errors.New("Invalid header for docker log")
	ErrInvalidData   = errors.New("Invalid data for docker log")
)

type LogEntry struct {
	Output  string
	Content string
}

func ReadAllDockerLogs(reader io.Reader) ([]LogEntry, error) {
	entries := make([]LogEntry, 0)

	for {
		entry, err := ReadOneDockerLog(reader)
		if err == nil {
			entries = append(entries, entry)
		} else if err == io.EOF {
			break
		} else {
			return entries, err
		}
	}
	return entries, nil
}

func ReadOneDockerLog(reader io.Reader) (LogEntry, error) {
	bufReader := bufio.NewReader(reader)
	entry := LogEntry{}

	header := make([]byte, 4)
	if _, err := io.ReadFull(bufReader, header); err != nil {
		return entry, err
	}

	output := "unknown"
	switch header[0] {
	case 0:
		output = "stdin"
	case 1:
		output = "stdout"
	case 2:
		output = "stderr"
	}

	var length uint32
	if err := binary.Read(bufReader, binary.BigEndian, &length); err != nil {
		return entry, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(bufReader, data); err != nil {
		return entry, err
	}
	entry.Output = output
	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	entry.Content = string(data)
	return entry, nil
}
