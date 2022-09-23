package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
)

var _ Logger = (*std)(nil)

type std struct {
	log  *log.Logger
	pool *sync.Pool
}

func NewStd(w io.Writer) Logger {
	return &std{
		log: log.New(w, "", 0),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (s *std) Log(level Level, kvs ...interface{}) error {
	lenKvs := len(kvs)
	if lenKvs == 0 {
		return nil
	}

	if (lenKvs & 1) == 1 {
		kvs = append(kvs, "log key value unpaired")
	}

	buf := s.pool.Get().(*bytes.Buffer)
	buf.WriteString(colorLevel(level))
	for i := 0; i < lenKvs; i += 2 {
		_, _ = fmt.Fprintf(buf, " %s=%v", withColor(level, kvs[i]), kvs[i+1])
	}
	_ = s.log.Output(4, buf.String())

	buf.Reset()
	s.pool.Put(buf)
	return nil
}

func (s *std) Close() error {
	return nil
}
