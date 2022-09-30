package form

import (
	"errors"

	"github.com/sado0823/go-kitx/pkg/encoding"
)

const Name = "form"

func init() {
	encoding.RegisterCodec(codec{})
}

type codec struct{}

func (c codec) Marshal(v interface{}) ([]byte, error) {
	return nil, errors.New("to be finished form encoding")
}

func (c codec) Unmarshal(data []byte, v interface{}) error {
	return errors.New("to be finished form encoding")
}

func (c codec) Name() string {
	return Name
}
