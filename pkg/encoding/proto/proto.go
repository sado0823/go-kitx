package proto

import (
	"github.com/sado0823/go-kitx/pkg/encoding"

	"google.golang.org/protobuf/proto"
)

const Name = "proto"

func init() {
	encoding.RegisterCodec(codec{})
}

type codec struct {
}

func (c codec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (c codec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (c codec) Name() string {
	return Name
}
