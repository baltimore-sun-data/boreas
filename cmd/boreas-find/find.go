package main

import (
	"fmt"
	"os"

	"github.com/baltimore-sun-data/boreas/cname"
)

func main() {
	if err := cname.FromArgs(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v", err)
		os.Exit(1)
	}
}
