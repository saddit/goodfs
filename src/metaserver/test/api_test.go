package test

import (
	"bytes"
	"common/cst"
	"common/hashslot"
	"common/proto/msg"
	"common/registry"
	"common/system"
	"common/util"
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"net/http"
	"testing"
	"time"
)

var (
	client = &http.Client{Timeout: 5 * time.Second}
	url    = "http://codespaces-409403:8010"
)

func TestClearEtcd(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := etcd.Delete(context.Background(), cst.EtcdPrefix.HashSlot, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cst.EtcdPrefix.HashSlot, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), cst.EtcdPrefix.Registry, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cst.EtcdPrefix.Registry, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), cst.EtcdPrefix.ObjectCap, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cst.EtcdPrefix.ObjectCap, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), cst.EtcdPrefix.ApiCredential, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cst.EtcdPrefix.ApiCredential, resp.Deleted)
	resp, err = etcd.Delete(context.Background(), cst.EtcdPrefix.SystemInfo, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cst.EtcdPrefix.SystemInfo, resp.Deleted)
}

func TestPostAPI(t *testing.T) {
	data := &msg.Metadata{
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
		Group:      "goodfs",
		Services:   []string{"metaserver"},
		ServerPort: "8080",
	})
	defer disc.MustRegister().Unregister()
	httpList := disc.GetServices("metaserver")
	rpcList := disc.GetServices("metaserver")
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
	key := cst.EtcdPrefix.ObjectCap
	resp, err := etcd.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cst.EtcdPrefix.ObjectCap, resp.Kvs)
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
	key := cst.EtcdPrefix.HashSlot
	resp, err := etcd.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		var i hashslot.SlotInfo
		_ = util.DecodeMsgp(&i, kv.Value)
		t.Logf("key=%s, value=%s", kv.Key, i.Slots)
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
	resp, err := etcd.Get(context.Background(), cst.EtcdPrefix.Registry, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		t.Logf("key=%s, value=%s", kv.Key, kv.Value)
	}
}

func TestGetSysInfo(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := etcd.Get(context.Background(), cst.EtcdPrefix.SystemInfo, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		var s system.Info
		_ = util.DecodeMsgp(&s, kv.Value)
		t.Logf("key=%s, value=%+v", kv.Key, s)
	}
}
