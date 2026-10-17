package quotas

// Total returns the sum of strictly positive values.
func Total(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}
