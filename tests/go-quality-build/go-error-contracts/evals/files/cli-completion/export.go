package lineexport

import (
	"bufio"
	"io"
)

func Export(r io.Reader, w io.WriteCloser) (int, error) {
	defer w.Close()
	scanner := bufio.NewScanner(r)
	n := 0
	for scanner.Scan() {
		if _, err := io.WriteString(w, scanner.Text()+"\n"); err != nil {
			return n, err
		}
		n++
	}
	return n, scanner.Err()
}
