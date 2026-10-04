package recordload_test

import (
	"errors"
	"fmt"
	"strings"

	"example.com/recordload"
)

func ExampleLoad() {
	records, err := recordload.Load(strings.NewReader("#settings\nhost=localhost\nbroken\nport=8080"))
	fmt.Println(records)
	var lineErr *recordload.LineError
	if errors.As(err, &lineErr) {
		fmt.Printf("line %d: %q\n", lineErr.Line, lineErr.Text)
	}
	// Output:
	// [{host localhost}]
	// line 3: "broken"
}
