package response

import (
	"encoding/json"
	"fmt"
	"io"
)

func MessageFromJSONBody(body io.ReadCloser) string {
	defer body.Close()
	bt, err := io.ReadAll(body)
	if err != nil {
		return err.Error()
	}
	mp := make(map[string]interface{})
	if err := json.Unmarshal(bt, &mp); err != nil {
		return err.Error()
	}
	return fmt.Sprint(mp["message"])
}

func IsOk(status int) bool {
	return status / 100 == 2
}

