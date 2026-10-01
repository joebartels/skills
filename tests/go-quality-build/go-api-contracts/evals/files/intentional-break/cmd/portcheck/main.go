package main

import (
	"example.com/portnum"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: portcheck PORT")
		os.Exit(2)
	}
	fmt.Println(portnum.Parse(os.Args[1]))
}
