package redis

import "context"

type Redis struct {
}

func New() *Redis {
	return &Redis{}
}

// TODO EvalCtx redis
func (s *Redis) EvalCtx(ctx context.Context, script string, keys []string, args ...interface{}) (val interface{}, err error) {
	panic("to be finished")
}
