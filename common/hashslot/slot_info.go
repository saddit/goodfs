package hashslot

//go:generate msgp

type SlotInfo struct {
	Location string   `json:"location" msg:"location"`
	Slots    []string `json:"slots" msg:"slots"`
	Peers    []string `json:"peers" msg:"peers"`
}
