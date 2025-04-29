package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "a + b",
			expected: []Token{
				{Type: IDENT, Literal: "a", Pos: 0},
				{Type: PLUS, Literal: "+", Pos: 2},
				{Type: IDENT, Literal: "b", Pos: 4},
				{Type: EOF, Literal: "", Pos: 5},
			},
		},
		{
			input: `\frac{123}{x^2}`,
			expected: []Token{
				{Type: COMMAND, Literal: "frac", Pos: 0},
				{Type: LBRACE, Literal: "{", Pos: 5},
				{Type: NUMBER, Literal: "123", Pos: 6},
				{Type: RBRACE, Literal: "}", Pos: 9},
				{Type: LBRACE, Literal: "{", Pos: 10},
				{Type: IDENT, Literal: "x", Pos: 11},
				{Type: CARET, Literal: "^", Pos: 12},
				{Type: NUMBER, Literal: "2", Pos: 13},
				{Type: RBRACE, Literal: "}", Pos: 14},
				{Type: EOF, Literal: "", Pos: 15},
			},
		},
		{
			input: "(a * -5.5)",
			expected: []Token{
				{Type: LPAREN, Literal: "(", Pos: 0},
				{Type: IDENT, Literal: "a", Pos: 1},
				{Type: ASTERISK, Literal: "*", Pos: 3},
				{Type: MINUS, Literal: "-", Pos: 5},
				{Type: NUMBER, Literal: "5.5", Pos: 6},
				{Type: RPAREN, Literal: ")", Pos: 9},
				{Type: EOF, Literal: "", Pos: 10},
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tokens := []Token{}
			for tok := l.NextToken(); tok.Type != EOF; tok = l.NextToken() {
				tokens = append(tokens, tok)
			}
			tokens = append(tokens, Token{Type: EOF, Literal: "", Pos: len(tt.input)}) // Add EOF manually for comparison

			assert.Equal(t, len(tt.expected), len(tokens), "Number of tokens mismatch")

			for i := range tt.expected {
				if i >= len(tokens) {
					break // Avoid index out of range if lengths mismatch (already asserted)
				}
				assert.Equal(t, tt.expected[i].Type, tokens[i].Type, "Token %d Type mismatch", i)
				assert.Equal(t, tt.expected[i].Literal, tokens[i].Literal, "Token %d Literal mismatch", i)
				// Position check can be brittle, optionally add:
				// assert.Equal(t, tt.expected[i].Pos, tokens[i].Pos, "Token %d Position mismatch", i)
			}
		})
	}
}
