package converter

import (
	"fmt"
)

// NOTE:
// - Must use little endian: binary.LittleEndian
// - File size must be 84 + (NumTriangles * 50)

const ( // In bytes
	HEADER_SIZE uint = 80
	FACET_SIZE = 50
	TRIANGLE_COUNT_SIZE = 4
	ATTRIBUTE_SIZE = 2
)

type Header [HEADER_SIZE]byte
type TriangleCount uint32
type Attribute uint16

type BinaryFacet struct {
	Facet

	Attrib Attribute
}

func CalculateFileSize(numTriangles uint32) uint {
	return uint(uint64(HEADER_SIZE) + uint64(TRIANGLE_COUNT_SIZE) + uint64(numTriangles)*uint64(FACET_SIZE))
}

// --- //

type BinarySTLObject struct {
	Header Header
	Triangles []BinaryFacet
}

func NewBinarySTLObject(header Header, triangleCount uint32) *BinarySTLObject {
	return &BinarySTLObject{
		Header: header,
		Triangles: make([]BinaryFacet, triangleCount),
	}
}

func (b *BinarySTLObject) GetHeader() Header { return b.Header }
func (b *BinarySTLObject) GetTriangleCount() uint32 { return uint32(len(b.Triangles)) }
func (b *BinarySTLObject) GetFacets() []BinaryFacet { return b.Triangles }

func (b *BinarySTLObject) WriteToHeader(data []byte, offset uint) error {
	if offset >= HEADER_SIZE || uint(len(data)) > HEADER_SIZE - offset {
		return fmt.Errorf("data with offset %d exceeds header size", offset)
	}
	copy(b.Header[offset:], data)
	return nil
}
