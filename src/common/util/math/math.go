package math

func LogicMod(a, b int) int {
	if a < 0 || b < 0 {
		return b + (a % b)
	}
	return a % b
}

func MinNumber[T Number](i, j T) T {
	if i < j {
		return i
	}
	return j
}

func MaxNumber[T Number](i, j T) T {
	if i > j {
		return i
	}
	return j
}

func MinInt(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func MaxInt(i, j int) int {
	if i < j {
		return j
	}
	return i
}

func MaxUint64(i, j uint64) uint64 {
	if i < j {
		return j
	}
	return i
}
