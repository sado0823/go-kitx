package json

import (
	"encoding/json"
	"reflect"

	"github.com/sado0823/go-kitx/pkg/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const Name = "json"

var (
	PbMarshalOption = protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	PbUnmarshalOption = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func init() {
	encoding.RegisterCodec(codec{})
}

type codec struct {
}

func (c codec) Marshal(v interface{}) ([]byte, error) {
	switch data := v.(type) {
	case json.Marshaler:
		return data.MarshalJSON()
	case proto.Message:
		return PbMarshalOption.Marshal(data)
	default:
		return json.Marshal(v)
	}
}

func (c codec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	case proto.Message:
		return PbUnmarshalOption.Unmarshal(data, m)
	default:
		rv := reflect.ValueOf(v)
		for rv := rv; rv.Kind() == reflect.Ptr; {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
			return PbUnmarshalOption.Unmarshal(data, m)
		}
		return json.Unmarshal(data, m)
	}
}

func (c codec) Name() string {
	return Name
}
