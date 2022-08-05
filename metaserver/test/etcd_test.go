package test

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestKeepAlive(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username: "root",
		Password: "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	ctx := context.Background()
	id, err := cli.Grant(ctx, 5)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.Put(ctx, "test8080", "123", clientv3.WithLease(id.ID))
	if err != nil {
		t.Fatal(err)
	}
	defer func ()  {
		cli.Revoke(ctx, id.ID)
		cli.Delete(ctx, "test8080")
	}()
	t.Log(resp.Header.String())
	ctx2, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	ch, err := cli.KeepAlive(ctx2, id.ID)
	if err != nil {
		t.Fatal(err)
	}
	for resp := range ch {
		t.Log("keep alive:", resp.String())
	}
	t.Log("close")
}