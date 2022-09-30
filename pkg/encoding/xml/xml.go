package xml


import (
	"encoding/xml"

	"github.com/sado0823/go-kitx/pkg/encoding"
)

const Name = "xml"

func init() {
	encoding.RegisterCodec(codec{})
}

type codec struct{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

func (codec) Name() string {
	return Name
}