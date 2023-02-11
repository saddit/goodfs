package mem

//go:generate msgp -tests=false #mem

type Status struct {
	All  uint64 `json:"all" msg:"all"`
	Used uint64 `json:"used" msg:"used"`
	Free uint64 `json:"free" msg:"free"`
	Self uint64 `json:"self" msg:"self"`
}
