package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <input.stl> <output.stl>\n", os.Args[0])
		os.Exit(2)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	fmt.Printf("Converting %s -> %s\n", inputPath, outputPath)
	fmt.Println("CLI implementation is scaffolded and ready for converter wiring")
	_ = outputPath
}
