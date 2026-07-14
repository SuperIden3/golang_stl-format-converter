package converter

import (
	"fmt"
	"strconv"
)

type Token uint

const (
	ILLEGAL Token = iota
	EOF
	SOLID
	NUMBER
	FACET
	NORMAL
	OUTER_LOOP // "outer loop" becomes one token
	VERTEX
	END_LOOP
	END_FACET
	END_SOLID
)

type NextToken_Return struct {
	Token Token
	Optional_SolidName SolidName
	Optional_Vector Vector
}

func TokenMappings() (map[string]Token, map[Token]string) {
	return map[string]Token{
		"solid": SOLID,
		"facet": FACET,
		"normal": NORMAL,
		"outer": OUTER_LOOP,
		"vertex": VERTEX,
		"endloop": END_LOOP,
		"endfacet": END_FACET,
		"endsolid": END_SOLID,
	}, map[Token]string{
		SOLID: "solid",
		FACET: "facet",
		NORMAL: "normal",
		OUTER_LOOP: "outer",
		VERTEX: "vertex",
		END_LOOP: "endloop",
		END_FACET: "endfacet",
		END_SOLID: "endsolid",
	}
}

func TokenList() []string {
	return []string{ "ILLEGAL", "EOF", "SOLID", "NUMBER", "FACET",  "NORMAL", "OUTER_LOOP", "VERTEX", "END_LOOP",  "END_FACET", "END_SOLID", }
}

func TokenFromString(s string) (Token, error) {
	mapping, _ := TokenMappings()
	if token, ok := mapping[s]; ok {
		return token, nil
	}
	if _, err := strconv.ParseFloat(s, 32); err == nil {
		return NUMBER, nil
	}
	return ILLEGAL, fmt.Errorf("Invalid token: %s", s)
}

func (t Token) String() (string, error) {
	_, reverseMapping := TokenMappings()
	if str, ok := reverseMapping[t]; ok {
		return str, nil
	}
	return "ILLEGAL", fmt.Errorf("Invalid token: %d", t)
}
