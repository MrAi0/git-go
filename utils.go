package main

import (
	"fmt"
	"os"
)

func errorPrintf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func must(err error) {
	if err != nil {
		errorPrintf("%s\n", err)
		os.Exit(1)
	}
}
