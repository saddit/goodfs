package test

import (
	"bytes"
	"common/cst"
	"common/hashslot"
	"common/registry"
	"common/system"
	"common/util"
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
	resp, err = etcd.Delete(context.Background(), cst.EtcdPrefix.PeersInfo, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cst.EtcdPrefix.PeersInfo, resp.Deleted)
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

func TestGetPeersInfo(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username:  "root",
		Password:  "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	key := cst.EtcdPrefix.PeersInfo
	resp, err := etcd.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		t.Logf("key=%s, value=%s", kv.Key, kv.Value)
	}
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
	resp, err := etcd.Get(context.Background(), cst.EtcdPrefix.HashSlot, clientv3.WithPrefix())
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
	resp, err := etcd.Get(context.Background(), cst.EtcdPrefix.Registry, clientv3.WithPrefix())
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
	resp, err := etcd.Get(context.Background(), cst.EtcdPrefix.SystemInfo, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		var sysInfo system.Info
		if err := util.DecodeMsgp(&sysInfo, kv.Value); err != nil {
			t.Fatal(err)
		}
		t.Logf("key=%s, value=%+v", kv.Key, sysInfo)
	}
}
