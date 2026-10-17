package main

import (
	"fmt"
	"io"
	"os"

	"example.org/streamwire"
)

func inspect(token string, out io.Writer) error {
	topic, offset, err := streamwire.DecodeCursor(token)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n%d\n", topic, offset)
	return err
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: streamctl TOKEN")
		os.Exit(1)
	}
	if err := inspect(os.Args[1], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
