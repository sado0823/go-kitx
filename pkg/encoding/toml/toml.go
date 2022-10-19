package toml

import (
	"github.com/sado0823/go-kitx/pkg/encoding"

	"github.com/pelletier/go-toml/v2"
)

// Name is the name registered for the yaml codec.
const Name = "toml"

func init() {
	encoding.RegisterCodec(codec{})
}

// codec is a Codec implementation with yaml.
type codec struct{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	return toml.Marshal(v)
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

func (codec) Name() string {
	return Name
}
