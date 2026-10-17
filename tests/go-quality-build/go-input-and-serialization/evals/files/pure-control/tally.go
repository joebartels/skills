package tally

// Total sums the odd values in a batch.
func Total(values []int) int {
	return sumOdd(values)
}

func sumOdd(values []int) int {
	total := 0
	for _, value := range values {
		if value%2 == 1 {
			total += value
		}
	}
	return total
}
