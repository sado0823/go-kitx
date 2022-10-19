package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sado0823/go-kitx/config"
	"github.com/sado0823/go-kitx/pkg/encoding"
	_ "github.com/sado0823/go-kitx/pkg/encoding/json"
	_ "github.com/sado0823/go-kitx/pkg/encoding/toml"
	_ "github.com/sado0823/go-kitx/pkg/encoding/yaml"
)

var _ config.Reader = (*file)(nil)

type (
	file struct {
		path     string
		ext      string
		name     string
		data     []byte
		fileInfo os.FileInfo
		withEnv  bool
	}

	Option func(file *file)
)

func WithEnv() Option {
	return func(file *file) {
		file.withEnv = true
	}
}

// New returns a config Reader with file , supported file(json,yaml,toml) and environment
func New(filepath string, opts ...Option) config.Reader {
	f := &file{path: filepath}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *file) Scan(v interface{}) error {
	codec := encoding.GetCodec(f.ext)
	if codec == nil {
		return fmt.Errorf("unsupported file extension:%q", f.ext)
	}
	return codec.Unmarshal(f.data, v)
}

func (f *file) Load() error {
	if err := f.load(); err != nil {
		return err
	}
	if f.withEnv {
		f.data = []byte(os.ExpandEnv(string(f.data)))
	}
	return nil
}

func (f *file) load() (err error) {
	var file *os.File
	file, err = os.Open(f.path)
	if err != nil {
		return err
	}
	defer file.Close()
	f.data, err = io.ReadAll(file)
	if err != nil {
		return err
	}

	f.fileInfo, err = file.Stat()
	if err != nil {
		return err
	}

	f.ext = filepath.Ext(f.path)
	if len(f.ext) > 1 {
		f.ext = f.ext[1:]
	}
	if len(f.ext) == 0 {
		return fmt.Errorf("invalid file extension %q", f.ext)
	}

	f.name = f.fileInfo.Name()

	return nil
}
