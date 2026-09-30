package limits

type LimitError struct{}

func (*LimitError) Error() string { return "limit must be between 1 and 100" }

func ValidateLimit(n int) error {
	var err *LimitError
	if n < 1 || n > 100 {
		err = &LimitError{}
	}
	return err
}
