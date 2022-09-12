package log

import "io"

type (
	WriterOption func(*writer)

	writer struct {
		helper *Helper
		level  Level
	}
)

func WithWriterMessageKey(msgKey string) WriterOption {
	return func(w *writer) {
		w.helper.msgKey = msgKey
	}
}

func WithWriterLevel(level Level) WriterOption {
	return func(w *writer) {
		w.level = level
	}
}

func NewWriter(l Logger, ops ...WriterOption) io.Writer {
	wt := &writer{
		helper: NewHelper(l, WithHelperMessageKey(DefaultMessageKey)),
		level:  LevelInfo,
	}

	for _, op := range ops {
		op(wt)
	}
	return wt
}

func (w *writer) Write(p []byte) (int, error) {
	w.helper.Log(w.level, w.helper.msgKey, string(p))
	return 0, nil
}
