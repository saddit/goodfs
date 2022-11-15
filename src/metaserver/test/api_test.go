package test

import (
	"bytes"
	"common/constrant"
	"common/hashslot"
	"common/registry"
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"metaserver/internal/entity"
	"net/http"
	"testing"
	"time"
)

var (
	client = &http.Client{Timeout: 5 * time.Second}
	url    = "http://codespaces-409403:8010"
)

func TestPostAPI(t *testing.T) {
	data := &entity.Metadata{
		Name: "test.txt",
	}
	bt, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Post(fmt.Sprintf("%s/metadata/test.txt", url), "application/json", bytes.NewBuffer(bt))
	if err != nil {
		t.Fatal(err)
	}
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error(string(res))
	} else {
		t.Log(string(res))
	}
}

func TestGetMeta(t *testing.T) {
	resp, err := client.Get(fmt.Sprintf("%s/metadata/test.txt?version=-1&date=1", url))
	if err != nil {
		t.Fatal(err)
	}
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}

func TestEtcdRegsitry(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	disc := registry.NewEtcdRegistry(etcd, registry.Config{
		Group:    "goodfs",
		Services: []string{"metaserver"},
		HttpAddr: "server-a:8080",
		RpcAddr:  "server-a:4040",
	})
	defer disc.MustRegister().Unregister()
	httpList := disc.GetServices("metaserver", false)
	rpcList := disc.GetServices("metaserver", true)
	t.Log(httpList, len(httpList))
	t.Log(rpcList, len(rpcList))
}

func TestGetObjectCaps(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	key := constrant.EtcdPrefix.ObjectCap
	resp, err := etcd.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.ObjectCap, resp.Kvs)
}

func TestGetHashSlot(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	key := constrant.EtcdPrefix.HashSlot
	resp, err := etcd.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.HashSlot, resp.Kvs)
}

func TestGetPeersInfo(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	key := constrant.EtcdPrefix.PeersInfo
	resp, err := etcd.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.PeersInfo, resp.Kvs)
}

func TestClearEtcd(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := etcd.Delete(context.Background(), constrant.EtcdPrefix.HashSlot, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.HashSlot, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), constrant.EtcdPrefix.PeersInfo, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.PeersInfo, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), constrant.EtcdPrefix.Registry, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.Registry, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), constrant.EtcdPrefix.ObjectCap, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.ObjectCap, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), constrant.EtcdPrefix.ApiCredential, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.ApiCredential, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), constrant.EtcdPrefix.SystemInfo, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(constrant.EtcdPrefix.SystemInfo, resp.Deleted)
}

func TestCalcHashSlot(t *testing.T) {
	input := "test123456.txt"
	output := hashslot.CalcBytesSlot([]byte(input))
	t.Log(output)
}

func TestGetSlots(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := etcd.Get(context.Background(), constrant.EtcdPrefix.HashSlot, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		t.Logf("key=%s, value=%s", kv.Key, kv.Value)
	}
}

func TestGetRegistry(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := etcd.Get(context.Background(), constrant.EtcdPrefix.Registry, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		t.Logf("key=%s, value=%s", kv.Key, kv.Value)
	}
}

func TestGetSystemInfo(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := etcd.Get(context.Background(), constrant.EtcdPrefix.SystemInfo, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		t.Logf("key=%s, value=%s", kv.Key, kv.Value)
	}
}
