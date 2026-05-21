package converter

type SolidName string
type Number float32 // Numbers in STL are 32-bit floats

type Vector struct {
	X, Y, Z Number
}

type Facet struct {
	Normal Vector
	Vertices [3]Vector
}
