package api

import "fmt"

func Health() string {
	return fmt.Sprintf("stl-format-converter API ready")
}

func ConvertSTL(inputFile, outputFile string) error {
}
