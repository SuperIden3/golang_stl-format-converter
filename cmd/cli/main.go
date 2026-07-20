package main

import (
	"fmt"
	"os"
	"superiden3.github.io/stl-format-converter/internal/api"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: %s <inputfile1> <outputfile1> [inputfile2] [outputfile2] [...] [inputfileN] [outputfileN]", os.Args[0])
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	for i := 1; i < len(os.Args); i += 2 {
		if !os.Exists(os.Args[i]) {
			fmt.Fprintln(os.Stderr, "Input file does not exist: %s, skipping next argument...", os.Args[i])
			continue
		}

		if i+1 >= len(os.Args) {
			fmt.Fprintln(os.Stderr, "Missing output file for input file: %s", os.Args[i])
			break
		}
	}
}
