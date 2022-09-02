package redis

import (
	"context"
	"fmt"

	rdsV8 "github.com/go-redis/redis/v8"

	"github.com/sado0823/go-kitx/kit/breaker"
)

const (
	Nil = rdsV8.Nil
)

type (
	Redis struct {
		Addr string
		Type string
		Pass string
		tls  bool
		brk  breaker.Breaker
	}
	Conn interface {
		rdsV8.Cmdable
	}
)

func New(addr string) *Redis {
	rds := &Redis{
		Addr: addr,
		Type: "node",
		Pass: "",
		tls:  false,
		brk:  breaker.New(),
	}
	err := rds.Ping(context.Background())
	if err != nil {
		panic(err)
	}
	return rds
}

func getConn(r *Redis) (Conn, error) {
	switch r.Type {
	case "cluster":
		return getCluster(r)
	case "node":
		return getClient(r)
	default:
		return nil, fmt.Errorf("invalid redis type: %s", r.Type)
	}
}

func acceptable(err error) bool {
	return err == nil || err == rdsV8.Nil
}

func (r *Redis) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (val interface{}, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := getConn(r)
		if err != nil {
			return err
		}
		val, err = conn.Eval(ctx, script, keys, args...).Result()
		return err
	}, acceptable)
	return val, err
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.brk.DoWithAcceptable(func() error {
		conn, err := getConn(r)
		if err != nil {
			return err
		}
		return conn.Ping(ctx).Err()
	}, acceptable)
}
