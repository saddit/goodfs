package credential


type emptExtral struct {
}

func (et *emptExtral) GetExtral() map[string][]string {
	return make(map[string][]string)
}