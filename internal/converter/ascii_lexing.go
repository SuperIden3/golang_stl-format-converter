package converter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Does file start with "solid"?
// ├── no → binary
// └── yes
//     ├── binary size matches?
//     │   ├── yes → binary
//     │   └── no
//     └── parse beginning as ASCII
//         ├── contains "facet normal"
//         ├── contains "outer loop"
//         └── contains "vertex"
//             yes → ASCII
//             no  → binary
func IsASCIIFile(file *os.File) (bool, error) {
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}
	defer func() {
		_, _ = file.Seek(currentPos, io.SeekStart)
	}()

	firstFive := make([]byte, 5)
	_, err = io.ReadFull(file, firstFive)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, err
	}

	if len(firstFive) >= 5 && strings.EqualFold(string(firstFive), "solid") {
		return true, nil
	}

	_, err = file.Seek(int64(HEADER_SIZE), io.SeekStart)
	if err != nil {
		return false, err
	}

	var triangleCount TriangleCount
	err = binary.Read(file, binary.LittleEndian, &triangleCount)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// The file is too small to be a valid binary STL header, so fall back to
			// ASCII keyword detection instead of rejecting it outright.
		} else {
			return false, err
		}
	} else {
		expectedSize := int64(CalculateFileSize(uint32(triangleCount)))
		fileSize, err := file.Seek(0, io.SeekEnd)
		if err != nil {
			return false, err
		}

		if fileSize == expectedSize {
			return false, nil
		}
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return false, err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return false, err
	}

	lower := strings.ToLower(string(content))
	hasFacetNormal := strings.Contains(lower, "facet normal")
	hasOuterLoop := strings.Contains(lower, "outer loop")
	hasVertex := strings.Contains(lower, "vertex")

	return hasFacetNormal && hasOuterLoop && hasVertex, nil
}

type ASCIILexer struct {
	scanner  *bufio.Scanner
	position Position
}

// New ASCIILexer from a given file object
func NewASCIILexer(file *os.File) *ASCIILexer {
	l := &ASCIILexer{
		scanner:  bufio.NewScanner(file),
		position: Position{Line: 1, Column: 1},
	}
	l.scanner.Split(scanWordsWithPosition(&l.position))
	return l
}

// "Close" the ASCIILexer by nulling the scanner
func (l *ASCIILexer) Close() {
	l.scanner = nil
	l.position = Position{Line: 0, Column: 0}
}

// Check if the ASCIILexer is closed
func (l *ASCIILexer) IsClosed() bool {
	return l.scanner == nil
}

// Get the current position in the file
func (l *ASCIILexer) GetPosition() Position {
	return l.position
}

// Custom split function that scans words and updates position
func scanWordsWithPosition(pos *Position) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		line := pos.Line
		column := pos.Column

		start := 0
		for start < len(data) && unicode.IsSpace(rune(data[start])) {
			if data[start] == '\n' {
				line++
				column = 1
			} else {
				column++
			}
			start++
		}

		if start == len(data) {
			if atEOF {
				pos.Line = line
				pos.Column = column
				return len(data), nil, nil
			}
			return 0, nil, nil
		}

		end := start
		for end < len(data) && !unicode.IsSpace(rune(data[end])) {
			column++
			end++
		}

		if end == len(data) && !atEOF {
			return 0, nil, nil
		}

		pos.Line = line
		pos.Column = column
		return end, data[start:end], nil
	}
}

// Returns the next float32
func (l *ASCIILexer) NextFloat32() (float32, error) {
	word, err := l.NextWord()
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(word, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

// Return the next word
func (l *ASCIILexer) NextWord() (string, error) {
	if l.IsClosed() {
		return "", fmt.Errorf("ASCIILexer is closed")
	}
	if !l.scanner.Scan() {
		if err := l.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return l.scanner.Text(), nil
}

func (l *ASCIILexer) NextToken() (NextToken_Return, error) {
	if l.IsClosed() {
		return NextToken_Return{}, fmt.Errorf("ASCIILexer is closed")
	}

	word, err := l.NextWord()
	if err != nil {
		if err == io.EOF {
			return NextToken_Return{Token: EOF}, nil
		}
		return NextToken_Return{}, err
	}

	token, err := TokenFromString(word)
	if err != nil {
		return NextToken_Return{}, err
	}

	switch token {
	case SOLID:
		solidName, err := l.NextWord()
		if err != nil {
			return NextToken_Return{}, err
		}
		return NextToken_Return{Token: SOLID, Optional_SolidName: SolidName(solidName)}, nil
	default:
		return NextToken_Return{Token: token}, nil
	}
}

func AddToken(token_arr []NextToken_Return, token NextToken_Return) []NextToken_Return {
	return append(token_arr, token)
}

func (l *ASCIILexer) ReadAll() ([]NextToken_Return, error) {
	if l.IsClosed() {
		return nil, fmt.Errorf("ASCIILexer is closed")
	}

	var tokens []NextToken_Return
	for {
		token, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		if token.Token == EOF {
			break
		}
		tokens = AddToken(tokens, token)
	}
	return tokens, nil
}
