package response

import (
	"common/logs"
	"encoding/json"
	"fmt"
	"io"
)

func MessageFromJSONBody(body io.ReadCloser) string {
	if body == nil {
		return "unknown"
	}
	defer body.Close()
	bt, err := io.ReadAll(body)
	if err != nil {
		logs.Std().Debugf("MessageFromJSONBody.ReadBody: %s", err)
		return "unknown"
	}
	mp := make(map[string]interface{})
	if err := json.Unmarshal(bt, &mp); err != nil {
		logs.Std().Debugf("MessageFromJSONBody.UnmarshalBody: %s", err)
		return "unknown"
	}
	return fmt.Sprint(mp["message"])
}

func IsOk(status int) bool {
	return status/100 == 2
}

func IsInternal(status int) bool {
	return status/100 == 5
}
