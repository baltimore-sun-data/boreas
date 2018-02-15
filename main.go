package main

import (
	"fmt"
	"os"

	"github.com/baltimore-sun-data/boreas/invalidator"
)

func main() {
	inv, err := invalidator.FromArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initial error: %v\n", err)
		os.Exit(3)
	}
	if err = inv.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Run time error: %v\n", err)
		os.Exit(1)
	}
}
