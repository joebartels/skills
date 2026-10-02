package csvselect

import (
	"encoding/csv"
	"io"
)

func WriteRows(w io.Writer, rows [][]string) error { return csv.NewWriter(w).WriteAll(rows) }
