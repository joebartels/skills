package clamp

// Clamp bounds value to the inclusive low/high interval.
func Clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
