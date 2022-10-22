package form

import (
	"net/url"

	"github.com/sado0823/go-kitx/pkg/encoding"
	"github.com/sado0823/go-kitx/pkg/encoding/form/scheme"

	"github.com/go-playground/form/v4"
	"google.golang.org/protobuf/proto"
)

const (
	NameFormData       = "form-data"
	NameFormUrlencoded = "x-www-form-urlencoded"
)

var (
	encoderForm = form.NewEncoder()
	decoderForm = scheme.NewDecoder()
)

func init() {
	decoderForm.SetAliasTag("json").IgnoreUnknownKeys(true).ZeroEmpty(true)
	encoderForm.SetTagName("json")
	encoding.RegisterCodec(formData{&base{encoder: encoderForm, decoder: decoderForm}})
	encoding.RegisterCodec(formUrlencoded{&base{encoder: encoderForm, decoder: decoderForm}})
}

type (
	formData struct {
		*base
	}
	formUrlencoded struct {
		*base
	}

	base struct {
		encoder *form.Encoder
		decoder *scheme.Decoder
	}
)

func (f formData) Name() string {
	return NameFormData
}

func (f formUrlencoded) Name() string {
	return NameFormUrlencoded
}

func (c base) Marshal(v interface{}) ([]byte, error) {
	var vs url.Values
	var err error
	if m, ok := v.(proto.Message); ok {
		vs, err = EncodeValues(m)
		if err != nil {
			return nil, err
		}
	} else {
		vs, err = c.encoder.Encode(v)
		if err != nil {
			return nil, err
		}
	}
	for k, v := range vs {
		if len(v) == 0 {
			delete(vs, k)
		}
	}
	return []byte(vs.Encode()), nil
}

func (c base) Unmarshal(data []byte, v interface{}) error {
	vs, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	return c.decoder.Decode(v, vs)
}

func (base) Name() string {
	return ""
}
