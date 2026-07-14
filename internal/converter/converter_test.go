package converter

import (
	"os"
	"testing"
)

func TestTokenFromStringRecognizesNumbers(t *testing.T) {
	tok, err := TokenFromString("1.23")
	if err != nil {
		t.Fatalf("expected number token but got error: %v", err)
	}
	if tok != NUMBER {
		t.Fatalf("expected NUMBER token, got %v", tok)
	}
}

func TestASCIILexerNextTokenHandlesNumbers(t *testing.T) {
	path := t.TempDir() + "/sample.stl"
	content := "solid test\nfacet\nnormal 0.1 0.2 0.3\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write sample file: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open sample file: %v", err)
	}
	defer file.Close()

	lexer := NewASCIILexer(file)
	defer lexer.Close()

	tok, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("expected first token but got error: %v", err)
	}
	if tok.Token != SOLID {
		t.Fatalf("expected SOLID token, got %v", tok.Token)
	}

	tok, err = lexer.NextToken()
	if err != nil {
		t.Fatalf("expected second token but got error: %v", err)
	}
	if tok.Token != FACET {
		t.Fatalf("expected FACET token, got %v", tok.Token)
	}

	tok, err = lexer.NextToken()
	if err != nil {
		t.Fatalf("expected third token but got error: %v", err)
	}
	if tok.Token != NORMAL {
		t.Fatalf("expected NORMAL token, got %v", tok.Token)
	}

	tok, err = lexer.NextToken()
	if err != nil {
		t.Fatalf("expected numeric token but got error: %v", err)
	}
	if tok.Token != NUMBER {
		t.Fatalf("expected NUMBER token, got %v", tok.Token)
	}
}
