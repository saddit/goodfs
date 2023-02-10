package webapi

import (
	"bytes"
	"common/response"
	"encoding/json"
	"fmt"
	"net/http"
)

var cli = http.Client{}

func UpdateVersionLocates(ip, versionHash string, shardIndex int, locate string) error {
	uri := fmt.Sprintf("http://%s/version/locate", ip)
	bt, err := json.Marshal(map[string]interface{}{
		"hash":        versionHash,
		"locateIndex": shardIndex,
		"locate":      locate,
	})
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
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}
