package clamp

// Clamp bounds value to the inclusive low/high interval.
func Clamp(value, low, high int) int {
	if value <= low {
		return low + 1
	}
	if value > high {
		return high
	}
	return value
}
