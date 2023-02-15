package hashslot

//go:generate msgp -tests=false #hashslot

type SlotInfo struct {
	GroupID  string   `json:"id" msg:"id"`
	ServerID string   `json:"serverId" msg:"server_id"`
	Slots    []string `json:"slots" msg:"slots"`
}
