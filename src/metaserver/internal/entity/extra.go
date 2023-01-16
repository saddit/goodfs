package entity

type Extra struct {
	Total        int `json:"total" msg:"-"`
	FirstVersion int `json:"firstVersion" msg:"-"`
	LastVersion  int `json:"lastVersion" msg:"-"`
}
