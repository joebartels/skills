package positive

func sumPositive(values []int) int {
	sum := 0
	for i := 0; i < len(values); i++ {
		if values[i] > 0 {
			sum += values[i]
		}
	}
	return sum
}
