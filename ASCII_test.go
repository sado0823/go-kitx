package kitx

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sado0823/go-kitx/transport/grpc"
	"github.com/sado0823/go-kitx/transport/http"
)

func Test_New(t *testing.T) {
	t.Log(logo)
}

func Test_NewApp(t *testing.T) {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		hs := http.NewServer(http.WithServerAddress("0.0.0.0:7001"))
		gs := grpc.NewServer(grpc.WithServerAddress("0.0.0.0:7002"))

		app := New(
			WithName("demo.app"),
			WithVersion("v0.0.00001"),
			WithMetadata(map[string]string{}),
			WithServer(hs, gs),
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

	client, err := grpc.DialInsecure(context.Background(),
		grpc.WithClientEndpoint("direct:///0.0.0.0:7002,0.0.0.0:7001"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(client.Target())

	err = client.Invoke(context.Background(), "/abc", 1, map[string]interface{}{})
	t.Log(err)
}
