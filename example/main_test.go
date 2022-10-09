package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sado0823/go-kitx"
	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/plugin/registry/etcd"
	"github.com/sado0823/go-kitx/transport/grpc"
	"github.com/sado0823/go-kitx/transport/http"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func Test_Log(t *testing.T) {

	log.Info("i am test")
}

func Test_NewApp(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"0.0.0.0:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var (
		wg     sync.WaitGroup
		discov = etcd.New(etcdClient)
	)

	wg.Add(1)
	go func() {
		hs := http.NewServer(http.WithServerAddress("0.0.0.0:7001"))
		gs := grpc.NewServer(grpc.WithServerAddress("0.0.0.0:7002"))

		app := kitx.New(
			kitx.WithName("demo.app"),
			kitx.WithVersion("v0.0.00001"),
			kitx.WithMetadata(map[string]string{}),
			kitx.WithServer(hs, gs),
			kitx.WithRegistrar(discov),
		)

		wg.Done()
		err := app.Run()
		if err != nil {
			t.Log(err)
			return
		}
	}()
	wg.Wait()
	time.Sleep(time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	client, err := grpc.DialInsecure(ctx,
		grpc.WithClientEndpoint("discovery:///demo.app"),
		//grpc.WithClientEndpoint("direct:///0.0.0.0:7002,0.0.0.0:7001"),
		grpc.WithClientDiscovery(discov),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(client.Target())

	for {
		err = client.Invoke(context.Background(), "/abc", 1, map[string]interface{}{})
		t.Log(err)
		time.Sleep(time.Second * 10)
	}

}
