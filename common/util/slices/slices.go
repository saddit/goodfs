package slices

func StringsReplace(arr []string, origin string, target string) bool {
	for i, a := range arr {
		if a == origin {
			arr[i] = target
			return true
		}
	}
	return false
}
