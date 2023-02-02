package webapi

import (
	"apiserver/internal/usecase/pool"
	"common/performance"
	"common/request"
	"common/response"
	"common/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func DeleteTmpObject(locate, id string) error {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionWrite, performance.KindOfHTTP, time.Since(st)) }()
	req, err := http.NewRequest(http.MethodDelete, tempRest(locate, id), nil)
	if err != nil {
		return err
	}
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func PostTmpObject(ip, name string, size int64) (string, error) {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionWrite, performance.KindOfHTTP, time.Since(st)) }()
	req, _ := http.NewRequest(http.MethodPost, tempRest(ip, name), nil)
	req.Header.Add("Size", fmt.Sprint(size))
	resp, e := pool.Http.Do(req)
	if e != nil {
		return "", e
	}
	res, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return "", fmt.Errorf("post temp object name=%v, return error response body, status=%v", name, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("post temp object name=%v, return code=%v, content=%s", name, resp.Status, string(res))
	}
	return string(res), nil
}

func PatchTmpObject(ip, id string, body io.Reader) error {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionWrite, performance.KindOfHTTP, time.Since(st)) }()
	req, _ := http.NewRequest(http.MethodPatch, tempRest(ip, id), body)
	keepAlive(req)
	resp, e := pool.Http.Do(req)
	if e != nil {
		return e
	}
	if resp.StatusCode == http.StatusBadRequest {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			content = util.StrToBytes(resp.Status)
		}
		return fmt.Errorf("patch temp object id=%v, return content=%v", id, string(content))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("patch temp object id=%v, return code=%v", id, resp.Status)
	}
	return nil
}

func PutTmpObject(ip, id string, compress bool) error {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionWrite, performance.KindOfHTTP, time.Since(st)) }()
	form := make(url.Values)
	form.Set("compress", fmt.Sprintf("%t", compress))
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprint(tempRest(ip, id), "?", form.Encode()), nil)
	keepAlive(req)
	resp, e := pool.Http.Do(req)
	if e != nil {
		return e
	}
	if resp.StatusCode == http.StatusBadRequest {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			content = util.StrToBytes(resp.Status)
		}
		return fmt.Errorf("put temp object id=%v, return content=%v", id, string(content))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("put temp object id=%v, return code=%v", id, resp.Status)
	}
	return nil
}

func HeadTmpObject(ip, id string) (int64, error) {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionRead, performance.KindOfHTTP, time.Since(st)) }()
	resp, e := http.Head(tempRest(ip, id))
	if e != nil {
		return 0, e
	}
	if resp.StatusCode == http.StatusBadRequest {
		if content, e := io.ReadAll(resp.Body); e == nil {
			return 0, fmt.Errorf("patch temp object id=%v, return content=%v", id, string(content))
		}
	}
	if resp.StatusCode == http.StatusNotFound {
		return 0, response.NewError(http.StatusNotFound, "object not found")
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("head temp object id=%v, return code=%v", id, resp.Status)
	}
	if str := resp.Header.Get("Size"); len(str) > 0 {
		size, e := strconv.ParseInt(str, 10, 0)
		if e != nil {
			return 0, fmt.Errorf("parse size string %s error: %v", str, e)
		}
		return size, nil
	}
	return 0, fmt.Errorf("response doesn't contains size")
}

func GetTmpObject(ip, name string, size int64) (*http.Response, error) {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionRead, performance.KindOfHTTP, time.Since(st)) }()
	req, err := http.NewRequest(http.MethodGet, tempRest(ip, name), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Size", util.ToString(size))
	return pool.Http.Do(req)
}

func GetObject(ip, name string, offset int, size int64, compress bool) (*http.Response, error) {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionRead, performance.KindOfHTTP, time.Since(st)) }()
	form := url.Values{}
	form.Set("compress", fmt.Sprintf("%t", compress))
	req, err := request.UrlValuesEncode(objectRest(ip, name), &form)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	req.Header.Set("Size", util.ToString(size))
	return pool.Http.Do(req)
}

func HeadObject(ip, id string) error {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionRead, performance.KindOfHTTP, time.Since(st)) }()
	resp, err := http.Head(objectRest(ip, id))
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return response.NewError(http.StatusNotFound, "object not found")
	}
	return fmt.Errorf("requset %s: %s", resp.Request.URL, resp.Status)
}

func PutObject(ip, id string, compress bool, body io.Reader) error {
	st := time.Now()
	defer func() { pool.Perform.PutAsync(performance.ActionWrite, performance.KindOfHTTP, time.Since(st)) }()
	form := url.Values{}
	form.Set("compress", fmt.Sprint(compress))
	req, err := http.NewRequest(http.MethodPut, fmt.Sprint(objectRest(ip, id), "?", form.Encode()), body)
	if err != nil {
		return err
	}
	keepAlive(req)
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("put object id=%v, return code=%v, %s", id, resp.Status, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func PingObject(ip string) error {
	resp, err := pool.Http.Get(fmt.Sprint("http://", ip, "/ping"))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func StatObject(ip string) (capacity int64, err error) {
	resp, err := pool.Http.Get(fmt.Sprint("http://", ip, "/stat"))
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
		return
	}
	capacity = util.ToInt64(resp.Header.Get("Capacity"))
	return
}

func objectRest(ip, id string) string {
	return fmt.Sprintf("http://%s/objects/%s", ip, id)
}

func tempRest(ip, id string) string {
	return fmt.Sprintf("http://%s/temp/%s", ip, id)
}

func keepAlive(req *http.Request) {
	req.Header.Set("Keep-Alive", "timeout=10, max=10000")
}
