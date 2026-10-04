package client

func SumPositive(values []int) int {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum
}
