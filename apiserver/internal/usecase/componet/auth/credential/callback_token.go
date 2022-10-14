package credential

import "encoding/json"

type CallbackToken struct {
	Bucket   string              `json:"bucket"`
	Region   string              `json:"region"`
	FileName string              `json:"file_name"`
	Version  int                 `json:"version"`
	Method   string              `json:"method"`
	Extra    map[string][]string `json:"-"`
}

func (ct *CallbackToken) GetUsername() string {
	bt, _ := json.Marshal(ct)
	return string(bt)
}

func (ct *CallbackToken) GetPassword() string {
	return ""
}

func (ct *CallbackToken) GetExtra() map[string][]string {
	return ct.Extra
}
