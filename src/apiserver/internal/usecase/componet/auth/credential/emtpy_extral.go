package credential

type emptyExtra struct {
}

func (et *emptyExtra) GetExtra() map[string][]string {
	return make(map[string][]string)
}
