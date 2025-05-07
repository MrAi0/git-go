package main

import (
	"fmt"
	"os"
)

func errorPrintf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}
