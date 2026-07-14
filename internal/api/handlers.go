package api

import "fmt"

func Health() string {
	return fmt.Sprintf("stl-format-converter API ready")
}
