package slices

import "common/util/math"

func StringsReplace(arr []string, origin string, target string) bool {
	for i, a := range arr {
		if a == origin {
			arr[i] = target
			return true
		}
	}
	return false
}

// SafeChunk [start, end], negative number means counting from tail to head
func SafeChunk[T any](arr []T, start, end int) []T {
	// a mod b = a mod (-b)
	// a mod b = - (-a mod b)

	if start < 0 {
		start = math.LogicMod(start, len(arr))
	} else if start >= len(arr) {
		start = len(arr) - 1
	}

	if end < 0 {
		end = math.LogicMod(end, len(arr))
	} else if end >= len(arr) {
		end = len(arr) - 1
	}

	if start > end {
		start, end = end, start
	}

	return arr[start : end+1]
}
