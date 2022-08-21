package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
