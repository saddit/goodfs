package response

import (
	"encoding/json"
	"io"
)

func MessageFromJSONBody(body io.ReadCloser) string {
	defer body.Close()
	bt, err := io.ReadAll(body)
	if err != nil {
		return err.Error()
	}
	mp := make(map[string]string)
	if err := json.Unmarshal(bt, &mp); err != nil {
		return err.Error()
	}
	return mp["message"]
}

