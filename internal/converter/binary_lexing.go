package converter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type BinaryLexer struct {
	       reader *bufio.Reader
	  tokenBuffer []NextToken_Return
	     position Position
	    triangles uint32
	trianglesRead uint32
	   headerRead bool
}

// New BinaryLexer from a given file object
func NewBinaryLexer(file *os.File) *BinaryLexer {
	l := &BinaryLexer{
		reader:   bufio.NewReader(file),
		position: Position{Line: 1, Column: 1},
	}
	return l
}

// "Close" the BinaryLexer by nulling the reader
func (l *BinaryLexer) Close() {
	l.reader      = nil
	l.position    = Position{Line: 0, Column: 0}
	l.tokenBuffer = nil
}

// Check if the BinaryLexer is closed
func (l *BinaryLexer) IsClosed() bool { return l.reader == nil }

// Get the current position in the file
func (l *BinaryLexer) GetPosition() Position {
	return l.position
}

// Returns the next float32
func (l *BinaryLexer) NextFloat32() (float32, error) {
	if l.IsClosed() { return 0, fmt.Errorf("BinaryLexer is closed") }

	var f float32
	err := binary.Read(l.reader, binary.LittleEndian, &f)
	if err != nil { return 0, err }

	return f, nil
}

// Read the 80-byte header and 4-byte triangle count
func (l *BinaryLexer) readHeaderAndCount() error {
	if l.headerRead { return nil }

	// Read 80 bytes header
	header := make([]byte, HEADER_SIZE)
	_, err := io.ReadFull(l.reader, header)
	if err != nil { return fmt.Errorf("failed to read binary STL header: %w", err) }

	// Read 4 bytes triangle count
	var count uint32
	err = binary.Read(l.reader, binary.LittleEndian, &count)
	if err != nil { return fmt.Errorf("failed to read binary STL triangle count: %w", err) }

	l.triangles = count
	l.headerRead = true

	// Seed with initial SOLID token containing header info
	l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{
		Token:              SOLID,
		Optional_SolidName: SolidName(string(header)),
	})

	return nil
}

// Parse the next triangle facet from binary stream and queue its structural tokens
func (l *BinaryLexer) parseNextFacetTokens() error {
	if l.trianglesRead >= l.triangles {
		l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{Token: END_SOLID})
		l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{Token: EOF})
		return nil
	}

	// 1. Read Normal Vector (3 x float32)
	nx, err := l.NextFloat32()
	if err != nil {
		return err
	}
	ny, err := l.NextFloat32()
	if err != nil {
		return err
	}
	nz, err := l.NextFloat32()
	if err != nil {
		return err
	}
	normal := Vector{X: Number(nx), Y: Number(ny), Z: Number(nz)}

	// 2. Read 3 Vertices (3 * 3 x float32)
	var vertices [3]Vector
	for i := 0; i < 3; i++ {
		vx, err := l.NextFloat32()
		if err != nil {
			return err
		}
		vy, err := l.NextFloat32()
		if err != nil {
			return err
		}
		vz, err := l.NextFloat32()
		if err != nil {
			return err
		}
		vertices[i] = Vector{X: Number(vx), Y: Number(vy), Z: Number(vz)}
	}

	// 3. Read Attribute Byte Count (uint16)
	var attr uint16
	err = binary.Read(l.reader, binary.LittleEndian, &attr)
	if err != nil {
		return err
	}

	// Increment processed triangle tracker
	l.trianglesRead++

	// 4. Translate the standard binary facet into corresponding sequential structural tokens
	l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{
		Token:           FACET,
		Optional_Vector: normal,
	})
	l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{Token: OUTER_LOOP})
	
	for i := 0; i < 3; i++ {
		l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{
			Token:           VERTEX,
			Optional_Vector: vertices[i],
		})
	}

	l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{Token: END_LOOP})
	l.tokenBuffer = append(l.tokenBuffer, NextToken_Return{Token: END_FACET})

	return nil
}

// NextToken gets the next structural STL token translating the binary data stream
func (l *BinaryLexer) NextToken() (NextToken_Return, error) {
	if l.IsClosed() {
		return NextToken_Return{}, fmt.Errorf("BinaryLexer is closed")
	}

	// Read initial header and count metadata if not already done
	if !l.headerRead {
		if err := l.readHeaderAndCount(); err != nil {
			return NextToken_Return{}, err
		}
	}

	// If the buffer is empty, extract structural tokens from the next facet block
	if len(l.tokenBuffer) == 0 {
		if err := l.parseNextFacetTokens(); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return NextToken_Return{Token: EOF}, nil
			}
			return NextToken_Return{}, err
		}
	}

	// Pop and return the front token of the queue buffer
	tok := l.tokenBuffer[0]
	l.tokenBuffer = l.tokenBuffer[1:]
	return tok, nil
}

// ReadAll iterates and processes all remaining tokens sequentially until EOF is reached
func (l *BinaryLexer) ReadAll() ([]NextToken_Return, error) {
	if l.IsClosed() {
		return nil, fmt.Errorf("BinaryLexer is closed")
	}

	var tokens []NextToken_Return
	for {
		tok, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		if tok.Token == EOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}
