package objectstream

import (
	"fmt"
	"goodfs/apiserver/global"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func DeleteTmpObject(locate, id string) {
	req, _ := http.NewRequest(http.MethodDelete, "http://"+locate+"/temp/"+id, nil)
	resp, e := global.Http.Do(req)
	if e != nil || resp.StatusCode != http.StatusOK {
		log.Println(e, resp.StatusCode)
	}
}

func PostTmpObject(ip, name string, size int64) (string, error) {
	req, _ := http.NewRequest(http.MethodPost, "http://"+ip+"/temp/"+name, nil)
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
	req, _ := http.NewRequest(http.MethodPatch, "http://"+ip+"/temp/"+id, body)
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
	req, _ := http.NewRequest(http.MethodPatch, "http://"+ip+"/temp/"+id, nil)
	req.Form.Add("name", name)
	resp, e := global.Http.Do(req)
	if e != nil {
		return e
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Put temp object id=%v, return code=%v", id, resp.Status)
	}
	return nil
}

func PutObject(ip, name string, body io.Reader) error {
	req, _ := http.NewRequest(http.MethodPut, "http://"+ip+"/objects/"+name, body)
	resp, e := global.Http.Do(req)
	if resp.StatusCode != http.StatusOK {
		e = fmt.Errorf("dataServer return http code %v", resp.StatusCode)
	}
	return e
}

func GetObject(ip, name string) (*http.Response, error) {
	return global.Http.Get("http://" + ip + "/objects/" + name)
}
