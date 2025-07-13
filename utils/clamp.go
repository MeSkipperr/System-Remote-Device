package utils

func Clamp(val, min, max int) int {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}
