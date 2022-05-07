package objectstream

import (
	"fmt"
	"goodfs/apiserver/global"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func DeleteTmpObject(locate, id string) {
	req, _ := http.NewRequest(http.MethodDelete, tempRest(locate, id), nil)
	resp, e := global.Http.Do(req)
	if e != nil || resp.StatusCode != http.StatusOK {
		log.Println(e, resp.StatusCode)
	}
}

func PostTmpObject(ip, name string, size int64) (string, error) {
	req, _ := http.NewRequest(http.MethodPost, tempRest(ip, name), nil)
	req.Header.Add("Size", fmt.Sprint(size))
	resp, e := global.Http.Do(req)
	if e != nil {
		return "", e
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Post temp object name=%v, return code=%v", name, resp.Status)
	}
	res, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return "", fmt.Errorf("Post temp object name=%v, return error response body", name)
	}
	return string(res), nil
}

func PatchTmpObject(ip, id string, body io.Reader) error {
	req, _ := http.NewRequest(http.MethodPatch, tempRest(ip, id), body)
	resp, e := global.Http.Do(req)
	if e != nil {
		return e
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Patch temp object id=%v, return code=%v", id, resp.Status)
	}
	return nil
}

func PutTmpObject(ip, id, name string) error {
	form := make(url.Values)
	form.Set("name", name)
	req, _ := http.NewRequest(http.MethodPut, tempRest(ip, id), strings.NewReader(form.Encode()))
	resp, e := global.Http.Do(req)
	if e != nil {
		return e
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Put temp object id=%v, return code=%v", id, resp.Status)
	}
	return nil
}

func GetObject(ip, name string) (*http.Response, error) {
	return global.Http.Get(objectRest(ip, name))
}

func objectRest(ip, id string) string {
	return fmt.Sprintf("http://%s/objects/%s", ip, id)
}

func tempRest(ip, id string) string {
	return fmt.Sprintf("http://%s/objects/%s", ip, id)
}
