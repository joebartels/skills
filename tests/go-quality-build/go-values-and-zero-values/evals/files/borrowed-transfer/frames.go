package framecollect

import "io"

type Source interface{ Next() ([]byte, error) }

func Collect(src Source) ([][]byte, error) {
	var out [][]byte
	for {
		frame, err := src.Next()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		var owned []byte
		if frame != nil {
			owned = make([]byte, len(frame))
			copy(owned, frame)
		}
		out = append(out, owned)
	}
}
