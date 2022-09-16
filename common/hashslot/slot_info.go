package hashslot

//go:generate msgp

type SlotInfo struct {
	Location string   `json:"location" msg:"location"`
	Checksum string   `json:"checksum" msg:"checksum"`
	Slots    []string `json:"slots" msg:"slots"`
	Peers    []string `json:"peers" msg:"peers"`
}
