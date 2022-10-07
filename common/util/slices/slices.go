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

func SafeChunk[T any](arr []T, start, end int) []T {
	if start < 0 && end < 0 {
		return arr
	} else if start < 0 {
		return arr[:end]
	} else if end < 0 {
		return arr[start:]
	} else if start >= len(arr) {
		return arr[0:0]
	} else if end > len(arr) {
		return arr[start:]
	}
	return arr[start:end]
}
