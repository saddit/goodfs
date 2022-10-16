package hashslot

//go:generate msgp

type SlotInfo struct {
	GroupID  string   `json:"id" msg:"id"`
	Location string   `json:"location" msg:"location"`
	Checksum string   `json:"checksum" msg:"checksum"`
	Slots    []string `json:"slots" msg:"slots"`
}
