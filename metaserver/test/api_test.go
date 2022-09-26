package test

import (
	"bytes"
	"common/constrant"
	"common/registry"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"metaserver/internal/entity"
	"net/http"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
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
	res, err := ioutil.ReadAll(resp.Body)
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
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}

func TestEtcdRegsitry(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username: "root",
		Password: "xianka",
	})
	if err != nil {
		t.Fatal(err)
	}
	disc := registry.NewEtcdDiscovery(etcd, &registry.Config{
		Group: "goodfs",
		Services: []string{"metaserver"},
	})
	
	for i := 0; i < 10; i++ {
		ls1 := disc.GetServicesWith("metaserver", true)
		ls2 := disc.GetServicesWith("metaserver", false)
		t.Log("masters", ls1, "slaves", ls2)
		time.Sleep(time.Second)
	}
	
}

func TestClearEtcd(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"pressed.top:2379"},
		Username: "root",
		Password: "xianka",
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
}