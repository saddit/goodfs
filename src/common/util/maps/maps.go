package maps

// OneOf returns first accessed key in this map through for-range
func OneOf[K, V comparable](mp map[K]V) (K, bool) {
	for k := range mp {
		return k, true
	}
	var zero K
	return zero, false
}

func Keys[K, V comparable](mp map[K]V) []K {
	var arr []K
	for k := range mp {
		arr = append(arr, k)
	}
	return arr
}

func Values[K, V comparable](mp map[K]V) []V {
	var arr []V
	for _, v := range mp {
		arr = append(arr, v)
	}
	return arr
}
