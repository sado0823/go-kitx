package redis

import (
	"crypto/tls"
	"io"

	"github.com/sado0823/go-kitx/pkg/syncx"

	rdsV8 "github.com/go-redis/redis/v8"
)

var clusterManager = syncx.NewResourceManager()

func getCluster(r *Redis) (*rdsV8.ClusterClient, error) {
	get, err := clusterManager.Get(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := rdsV8.NewClusterClient(&rdsV8.ClusterOptions{
			Addrs:        []string{r.Addr},
			Password:     r.Pass,
			MaxRetries:   3,
			MinIdleConns: 8,
			TLSConfig:    tlsConfig,
		})
		return store, nil
	})
	if err != nil {
		return nil, err
	}

	return get.(*rdsV8.ClusterClient), nil
}

var clientManager = syncx.NewResourceManager()

func getClient(r *Redis) (*rdsV8.Client, error) {
	get, err := clientManager.Get(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := rdsV8.NewClient(&rdsV8.Options{
			Addr:         r.Addr,
			Password:     r.Pass,
			DB:           0,
			MaxRetries:   3,
			MinIdleConns: 8,
			TLSConfig:    tlsConfig,
		})
		return store, nil
	})
	if err != nil {
		return nil, err
	}

	return get.(*rdsV8.Client), nil
}
