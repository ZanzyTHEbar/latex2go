package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// TokenType identifies the type of lexed token.
type TokenType int

// Token represents a lexed token (type, literal value, position).
type Token struct {
	Type    TokenType
	Literal string
	Pos     int // Starting position of the token in the input string
}

// Define token types.
const (
	ILLEGAL TokenType = iota // Illegal token (error condition)
	EOF                      // End of File

	// Literals
	IDENT  // Identifier (variable or function name like 'x', 'sin')
	NUMBER // Numeric literal (e.g., 3.14, 42)

	// Operators
	PLUS       // +
	MINUS      // -
	ASTERISK   // *
	SLASH      // /
	CARET      // ^
	EQUALS     // =
	EXCLAMATION// ! (factorial)

	// Delimiters
	LPAREN     // (
	RPAREN     // )
	LBRACE     // {
	RBRACE     // }
	UNDERSCORE // _

	// LaTeX Commands (treated specially)
	COMMAND    // e.g., \frac, \sqrt, \sin
	BEGIN      // \begin{...}
	END        // \end{...}
)

// Lexer holds the state of the scanner.
type Lexer struct {
	input        string // Input string being scanned
	position     int    // Current position in input (points to current char)
	readPosition int    // Current reading position in input (after current char)
	ch           rune   // Current char under examination
}

// NewLexer creates a new Lexer instance.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // Initialize l.ch, l.position, l.readPosition
	return l
}

// readChar gives us the next character and advances our position in the input string.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for "NUL", signifies EOF or not read yet
	} else {
		var size int
		l.ch, size = utf8.DecodeRuneInString(l.input[l.readPosition:])
		if l.ch == utf8.RuneError && size == 1 {
			l.ch = '?'
		}
	}
	l.position = l.readPosition
	l.readPosition += utf8.RuneLen(l.ch)
}

// peekChar looks ahead at the next character without consuming it.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// NextToken scans the input and returns the next token.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Pos = l.position

	switch l.ch {
	case '+':
		tok = newToken(PLUS, l.ch)
	case '-':
		tok = newToken(MINUS, l.ch)
	case '*':
		tok = newToken(ASTERISK, l.ch)
	case '/':
		tok = newToken(SLASH, l.ch)
	case '^':
		tok = newToken(CARET, l.ch)
	case '=':
		tok = newToken(EQUALS, l.ch)
	case '!':
		tok = newToken(EXCLAMATION, l.ch)
	case '_':
		tok = newToken(UNDERSCORE, l.ch)
	case '(':
		tok = newToken(LPAREN, l.ch)
	case ')':
		tok = newToken(RPAREN, l.ch)
	case '{':
		tok = newToken(LBRACE, l.ch)
	case '}':
		tok = newToken(RBRACE, l.ch)
	case '\\':
		tok.Type = COMMAND
		cmdStr := l.readCommand()
		tok.Literal = cmdStr
		tok.Pos = l.position
		
		// Special handling for \begin and \end
		if cmdStr == "begin" {
			tok.Type = BEGIN
		} else if cmdStr == "end" {
			tok.Type = END
		}
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = IDENT
			return tok
		} else if isDigit(l.ch) {
			tok.Type = NUMBER
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

// skipWhitespace consumes whitespace characters.
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readCommand() string {
	position := l.position + 1
	l.readChar()
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	hasDecimal := false
	for isDigit(l.ch) || (l.ch == '.' && !hasDecimal) {
		if l.ch == '.' {
			if !isDigit(l.peekChar()) {
				break
			}
			hasDecimal = true
		}
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType TokenType, ch rune) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}

func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case NUMBER:
		return "NUMBER"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case ASTERISK:
		return "ASTERISK"
	case SLASH:
		return "SLASH"
	case CARET:
		return "CARET"
	case EQUALS:
		return "EQUALS"
	case EXCLAMATION:
		return "EXCLAMATION"
	case UNDERSCORE:
		return "UNDERSCORE"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case COMMAND:
		return "COMMAND"
	case BEGIN:
		return "BEGIN"
	case END:
		return "END"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(t))
	}
}
