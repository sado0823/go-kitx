package etcd

import (
	"testing"
)

func Test_New(t *testing.T) {
	//client, err := clientv3.New(clientv3.Config{
	//	Endpoints: []string{"0.0.0.0:2379"},
	//})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//var (
	//	ctx = context.Background()
	//)
	//grant, err := client.Grant(ctx, int64((time.Minute * 5).Seconds()))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//_, err = client.Put(ctx, "/a/b/c", "value666", clientv3.WithLease(grant.ID))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//kv := clientv3.NewKV(client)
	//watch := clientv3.NewWatcher(client)
	//watchChan := watch.Watch(ctx, "/a/b", clientv3.WithPrefix(), clientv3.WithRev(0), clientv3.WithKeysOnly())
	//err = watch.RequestProgress(context.Background())
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//t.Log(grant.ID)
	//
	//select {
	//case <-ctx.Done():
	//	t.Fatal(ctx.Err())
	//case <-watchChan:
	//	resp, err := kv.Get(ctx, "/a/b", clientv3.WithPrefix())
	//	if err != nil {
	//		t.Fatal(ctx.Err())
	//	}
	//	for _, kv := range resp.Kvs {
	//		t.Logf("kv name: %s, value:%s", kv.Key, kv.Value)
	//	}
	//}
	//
	//keepAlive, err := client.KeepAlive(ctx, grant.ID)
	//_, ok := <-keepAlive
	//t.Log(keepAlive, err, ok)

	//select {
	//case <-ctx.Done():
	//	t.Fatal(ctx.Err())
	//case <-watchChan:
	//	resp, err := kv.Get(ctx, "/a/b", clientv3.WithPrefix())
	//	if err != nil {
	//		t.Fatal(ctx.Err())
	//	}
	//	for _, kv := range resp.Kvs {
	//		t.Logf("kv name: %s, value:%s", kv.Key, kv.Value)
	//	}
	//}

}
