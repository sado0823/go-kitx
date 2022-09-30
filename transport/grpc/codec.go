package grpc

import (
	"fmt"

	"github.com/sado0823/go-kitx/pkg/encoding"
	"github.com/sado0823/go-kitx/pkg/encoding/json"

	googleEncoding "google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
)

func init() {
	googleEncoding.RegisterCodec(codec{})
}

// codec is a Codec implementation with protobuf. It is the default codec for gRPC.
type codec struct{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	vv, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
	return encoding.GetCodec(json.Name).Marshal(vv)
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	vv, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
	return encoding.GetCodec(json.Name).Unmarshal(data, vv)
}

func (codec) Name() string {
	return json.Name
}
