package util

import (
	"encoding/json"
	"net/http"

	"github.com/golang/protobuf/proto"
)

type Marshaler func(interface{}) ([]byte, error)
type Unmarshaler func([]byte, interface{}) error

func GetMarshalers(r *http.Request) (Marshaler, Unmarshaler) {
	if r.URL.Query().Get("method") == "web" {
		return json.Marshal, json.Unmarshal
	}
	return protoMarshalFunc, protoUnmarshalFunc
}

// Adapters
func protoMarshalFunc(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func protoUnmarshalFunc(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}
