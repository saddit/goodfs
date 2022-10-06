package webapi

import (
	"bytes"
	"common/pb"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var cli = http.Client{}

func VersionsByHash(ip, hash string) ([]*pb.Version, error) {
	resp, err := cli.Get(fmt.Sprintf("http://%s/version/list?hash=%s", ip, hash))
	if err != nil {
		return nil, err
	}
	bt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res []*pb.Version
	if err = json.Unmarshal(bt, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func UpdateVersionLocates(ip, name string, version int, locates []string) error {
	uri := fmt.Sprintf("http://%s/metadata_version/%s/locates?version=%d", ip, name, version)
	bt, err := json.Marshal(map[string]interface{}{"locations": locates})
	if err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(bt))
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("update locates fail: status=" + resp.Status)
	}
	return nil
}
