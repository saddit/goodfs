package math

func LogicMod(a, b int) int {
	if a < 0 || b < 0 {
		return b + (a % b)
	}
	return a % b
}