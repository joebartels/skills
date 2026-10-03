package positive

func sumPositive(values []int) int {
	sum := 0
	for _, value := range values {
		if value > 0 {
			sum += value
		}
	}
	return sum
}
