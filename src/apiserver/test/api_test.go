package test

import (
	"bytes"
	"common/datasize"
	"common/response"
	"common/util/crypto"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const (
	baseUrl = "https://localhost:8080"
)

var (
	cli = http.Client{}
)

func TestBigUpload(t *testing.T) {
	totalSize := datasize.MB * 12
	data := make([]byte, totalSize)
	for i := range data {
		data[i] = 'D'
	}
	// init
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprint(baseUrl, "/v1/big/testBig-1.5.bytes"), nil)
	req.Header.Set("Digest", crypto.SHA256(data))
	req.Header.Set("Size", fmt.Sprint(len(data)))
	req.Header.Set("Bucket", "test-1")
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusOK {
		bt, _ := io.ReadAll(resp.Body)
		t.Logf("finish upload since file duplicate:\n %s", string(bt))
		return
	}
	if resp.StatusCode != http.StatusCreated {
		bt, _ := io.ReadAll(resp.Body)
		t.Fatal(string(bt))
	}
	uploadPath := resp.Header.Get("Location")
	minPartSize := datasize.MustParse(resp.Header.Get("Min-Part-Size")) * 2
	var cursor datasize.DataSize
	//// first time: upload 3.5 MB and abort
	for cursor < datasize.MB*3+datasize.KB*512 {
		end := cursor + minPartSize
		if end > totalSize {
			end = totalSize
		}
		req, _ = http.NewRequest(http.MethodPatch, fmt.Sprint(baseUrl, uploadPath), bytes.NewBuffer(data[cursor:end]))
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", cursor, end))
		resp, err = cli.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusPartialContent {
			bt, _ := io.ReadAll(resp.Body)
			t.Fatal(string(bt))
		}
		t.Logf("current size from server: %s", datasize.DataSize(resp.ContentLength))
		cursor += minPartSize
	}
	t.Log("first time upload success")
	// recover upload
	resp, err = cli.Head(fmt.Sprint(baseUrl, uploadPath))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(io.ReadAll(resp.Body))
	}
	cursor = datasize.DataSize(resp.ContentLength)
	t.Logf("reset cursor to %d", cursor)
	// second time: upload all
	for cursor < totalSize {
		end := cursor + minPartSize
		if end > totalSize {
			end = totalSize
		}
		req, _ = http.NewRequest(http.MethodPatch, fmt.Sprint(baseUrl, uploadPath), bytes.NewBuffer(data[cursor:end]))
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", cursor, end))
		resp, err = cli.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if !response.IsOk(resp.StatusCode) {
			bt, _ := io.ReadAll(resp.Body)
			t.Fatal(string(bt))
		}
		t.Logf("current size from server: %s", datasize.DataSize(resp.ContentLength))
		cursor += minPartSize
	}
	t.Log("second time upload success")
	bt, _ := io.ReadAll(resp.Body)
	t.Log(string(bt))
}
