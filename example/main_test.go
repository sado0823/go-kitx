package main

import (
	"context"
	"fmt"
	http2 "net/http"
	"sync"
	"testing"
	"time"

	"github.com/sado0823/go-kitx"
	"github.com/sado0823/go-kitx/errorx"
	"github.com/sado0823/go-kitx/internal/test/pbhelloworld"
	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/plugin/registry/etcd"
	"github.com/sado0823/go-kitx/transport/grpc"
	"github.com/sado0823/go-kitx/transport/http"
	"github.com/sado0823/go-kitx/transport/pbchain"

	clientv3 "go.etcd.io/etcd/client/v3"
)


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
		hs := http.NewServer(
			http.WithServerAddress("0.0.0.0:7001"),
			http.WithServerFilter(func(handler http2.Handler) http2.Handler {
				return http2.HandlerFunc(func(w http2.ResponseWriter, r *http2.Request) {
					log.Info("in server filter1")
					handler.ServeHTTP(w, r)
					log.Info("in server filter1 after")
				})
			}, func(handler http2.Handler) http2.Handler {
				return http2.HandlerFunc(func(w http2.ResponseWriter, r *http2.Request) {
					log.Info("in server filter2")
					handler.ServeHTTP(w, r)
					log.Info("in server filter2 after")
				})
			}),
			http.WithServerPBChain(
				pbchain.LoggingServer(log.GetGlobal()),
				pbchain.Validator(),
			),
		)

		r := hs.Route("/")
		pbhelloworld.RegisterGreeterHTTPServer(r, pbhelloworld.UnimplementedGreeterServer{})

		r.GET("/ping", func(c http.Context) error {
			return c.JSON(200, map[string]interface{}{
				"path": "ping",
			})
		})

		r.GET("/error", func(c http.Context) error {
			return errorx.BadRequest("错了错了", "msg").WithMetadata(map[string]string{
				"a": "b",
			}).WithCause(fmt.Errorf("got err, path:%s", c.Request().URL))
		})

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

	//for {
	err = client.Invoke(context.Background(), "/abc", 1, map[string]interface{}{})
	t.Log(err)
	//time.Sleep(time.Second * 10)
	//}

	httpClient, err := http.NewClient(ctx, http.WithClientEndpoint("0.0.0.0:7001"))
	if err != nil {
		t.Fatal(err)
	}

	args := map[string]interface{}{}
	reply := map[string]interface{}{}
	//http.ContentType("application/json")
	err = httpClient.Invoke(ctx, "get", "/ping", &args, &reply)
	t.Log(err)
	t.Log(args)
	t.Log(reply)

	args2 := map[string]interface{}{}
	reply2 := map[string]interface{}{}
	err = httpClient.Invoke(ctx, "post", "/add/user", &args2, &reply2)
	t.Log(err)
	t.Log(args2)
	t.Log(reply2)
}
