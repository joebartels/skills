package cursorload

type Rows interface {
	Next() bool
	Scan(*string) error
	Err() error
	Close() error
}

func Collect(rows Rows) ([]string, error) {
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return out, err
		}
		out = append(out, s)
	}
	return out, nil
}
