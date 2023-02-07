package maps

// OneOf returns first access key in this map by for-range
func OneOf[K comparable, V comparable](mp map[K]V) (K, bool) {
	for k := range mp {
		return k, true
	}
	var zero K
	return zero, false
}
