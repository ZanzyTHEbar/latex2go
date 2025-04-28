package parser

import (
	"fmt"
	"strconv"
	"strings" // Needed now for function names

	internalast "github.com/placeholderuser/latex2go/internal/domain/ast"
)

// --- Operator Precedence ---
const (
	_ int = iota
	LOWEST
	SUM      // +, -
	PRODUCT  // *, /
	EXPONENT // ^
	PREFIX   // -X or !X (if we add negation)
	CALL     // myFunction(X) or \command{X}
)

var precedences = map[TokenType]int{
	PLUS:     SUM,
	MINUS:    SUM,
	ASTERISK: PRODUCT,
	SLASH:    PRODUCT,
	CARET:    EXPONENT,
	LPAREN:   CALL, // For grouped expressions like (a+b)
	// LBRACE:   CALL, // No longer needed as infix/prefix operator itself
	COMMAND:  CALL, // For commands like \sin x (though args usually use braces)
}

// --- Parser Implementation ---

// Function types for Pratt parsing
type (
	prefixParseFn func() (internalast.Expr, error)
	infixParseFn  func(internalast.Expr) (internalast.Expr, error)
)

// Parser holds the lexer, tokens, errors, and parsing functions.
type Parser struct {
	l      *Lexer
	errors []string

	curToken  Token
	peekToken Token

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

// NewParser creates a new parser instance (used by the service).
// The actual lexer/state is created within the Parse method.
func NewParser() *Parser {
	return &Parser{}
}

// newStatefulParser creates a new stateful Parser instance for a specific parsing run.
func newStatefulParser(l *Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[TokenType]prefixParseFn),
		infixParseFns:  make(map[TokenType]infixParseFn),
	}

	// Register parsing functions
	p.registerPrefix(IDENT, p.parseIdentifier)
	p.registerPrefix(NUMBER, p.parseNumberLiteral)
	p.registerPrefix(LPAREN, p.parseGroupedExpression)
	p.registerPrefix(MINUS, p.parsePrefixExpression)    // For unary minus
	p.registerPrefix(COMMAND, p.parseCommandExpression) // For \sqrt, \frac, etc.

	p.registerInfix(PLUS, p.parseInfixExpression)
	p.registerInfix(MINUS, p.parseInfixExpression)
	p.registerInfix(ASTERISK, p.parseInfixExpression)
	p.registerInfix(SLASH, p.parseInfixExpression)
	p.registerInfix(CARET, p.parseInfixExpression)
	// p.registerInfix(LBRACE, p.parseCommandArguments) // REMOVED - Command args handled in parseCommandExpression

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns any parsing errors encountered.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.errors = append(p.errors, fmt.Sprintf("parse error at pos %d: %s", p.curToken.Pos, msg))
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseExpression is the main entry point for parsing the entire expression.
// It replaces the old placeholder Parse method.
func (p *Parser) ParseExpression() (internalast.Expr, error) {
fmt.Printf("ParseExpression START. cur=%s peek=%s\n", p.curToken.Type, p.peekToken.Type) // DEBUG
expr, err := p.parseExpression(LOWEST)
if err != nil {
// Add the error if it's not already captured (some funcs might add directly)
		// For simplicity, just return it here. Refine error aggregation if needed.
		return nil, err
	}

	// Check if any errors were collected during parsing
	if len(p.errors) > 0 {
		// Combine errors into a single error message
		return nil, fmt.Errorf("parsing failed:\n\t%s", strings.Join(p.errors, "\n\t"))
	}

	// Ensure we consumed the whole input (optional, depends on requirements)
	if p.peekToken.Type != EOF {
		p.addError("unexpected token '%s' after expression", p.peekToken.Literal)
		return nil, fmt.Errorf(p.errors[len(p.errors)-1]) // Return the last error
	}

	return expr, nil
}

// --- Pratt Parsing Core ---

func (p *Parser) parseExpression(precedence int) (internalast.Expr, error) {
fmt.Printf(" -> parseExpression(%d). cur=%s peek=%s\n", precedence, p.curToken.Type, p.peekToken.Type) // DEBUG
prefix := p.prefixParseFns[p.curToken.Type]
if prefix == nil {
err := fmt.Errorf("no prefix parse function found for token %s ('%s')", p.curToken.Type, p.curToken.Literal)
		p.addError(err.Error())
		return nil, err
	}
	leftExp, err := prefix()
	if err != nil {
		return nil, err // Error already added by prefix func or above
	}

	for p.peekToken.Type != EOF && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			// Not an infix operator we handle, or end of this precedence level
			return leftExp, nil
		}

		p.nextToken() // Consume the infix operator

		leftExp, err = infix(leftExp) // Pass the left expression to the infix function
		if err != nil {
			return nil, err // Error added by infix func
		}
	}

	return leftExp, nil
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// --- Parsing Functions ---

func (p *Parser) parseIdentifier() (internalast.Expr, error) {
	// Assumes single-letter variables for now, as per lexer
	return &internalast.Variable{Name: p.curToken.Literal}, nil
}

func (p *Parser) parseNumberLiteral() (internalast.Expr, error) {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		err = fmt.Errorf("could not parse '%s' as float: %w", p.curToken.Literal, err)
		p.addError(err.Error())
		return nil, err
	}
	return &internalast.NumberLiteral{Value: val}, nil
}

func (p *Parser) parsePrefixExpression() (internalast.Expr, error) {
	// Currently only handles unary minus
	if p.curToken.Type != MINUS {
		err := fmt.Errorf("expected prefix operator (e.g., '-'), got %s", p.curToken.Type)
		p.addError(err.Error())
		return nil, err
	}

	// Treat unary minus as multiplication by -1 for simplicity in AST/Generator
	// Alternatively, add UnaryExpr to internal AST
	p.nextToken() // Consume the '-'
	rightExpr, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}

	return &internalast.BinaryExpr{
		Op:    "*",
		Left:  &internalast.NumberLiteral{Value: -1.0},
		Right: rightExpr,
	}, nil

	// // --- Alternative: Using UnaryExpr ---
	// // Requires adding UnaryExpr struct to internalast
	//
	//	expr := &internalast.UnaryExpr{
	//	    Operator: p.curToken.Literal,
	//	}
	//
	// p.nextToken() // Consume the operator
	// var err error
	// expr.Right, err = p.parseExpression(PREFIX)
	//
	//	if err != nil {
	//	    return nil, err
	//	}
	//
	// return expr, nil
	// // --- End Alternative ---
}

func (p *Parser) parseInfixExpression(left internalast.Expr) (internalast.Expr, error) {
	expr := &internalast.BinaryExpr{
		Op:   p.curToken.Literal,
		Left: left,
	}
	precedence := p.curPrecedence()
	p.nextToken() // Consume the operator
	var err error
	expr.Right, err = p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parseGroupedExpression() (internalast.Expr, error) {
	p.nextToken() // Consume '('
	expr, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if !p.expectPeek(RPAREN) {
		// Error already added by expectPeek
		return nil, fmt.Errorf("missing closing parenthesis")
	}
	return expr, nil
}

// parseCommandExpression handles commands like \sqrt{x}, \frac{a}{b}, etc.
// It assumes arguments are enclosed in braces {} immediately following the command.
func (p *Parser) parseCommandExpression() (internalast.Expr, error) {
	// Current token is COMMAND
	funcName := p.curToken.Literal
	args := []internalast.Expr{}

	// Loop while the *next* token is an opening brace for an argument
	for p.peekToken.Type == LBRACE {
		p.nextToken() // Consume COMMAND token (first iteration) or RBRACE (subsequent)
		// Now curToken is LBRACE.

		// We are at LBRACE, consume it to get to the argument start
		p.nextToken() // Consume LBRACE. curToken is now first token of arg.

		argExpr, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, argExpr)

		// After parseExpression, curToken is the last token of the arg.
		// We expect peekToken to be RBRACE.
		if p.peekToken.Type != RBRACE {
			p.peekError(RBRACE)
			return nil, fmt.Errorf("missing '}' after argument for command \\%s", funcName)
		}
		p.nextToken() // Consume RBRACE. curToken is now RBRACE.
		// The loop condition checks peekToken for the next LBRACE.
	}

	// Check if any arguments were parsed
	if len(args) == 0 {
		// Handle commands without braces if needed (e.g., \sin x) - currently unsupported
		err := fmt.Errorf("expected '{' arguments after command '\\%s', got %s", funcName, p.peekToken.Type)
		p.addError(err.Error())
		return nil, err
	}

	// --- Argument Count Validation ---
	requiredArgs := -1 // -1 means variable or unknown
	switch strings.ToLower(funcName) {
	case "frac":
		requiredArgs = 2
	case "sqrt", "sin", "cos", "tan": // Add other single-arg functions here
		requiredArgs = 1
	}

	if requiredArgs != -1 && len(args) != requiredArgs {
		err := fmt.Errorf("\\%s requires %d argument(s), got %d", funcName, requiredArgs, len(args))
		p.addError(err.Error())
		return nil, err
	}
	// --- End Validation ---

	return &internalast.FuncCall{
		FuncName: funcName,
		Args:     args,
	}, nil
}

// parseCommandArguments removed, logic integrated into parseCommandExpression

// expectPeek checks the type of the next token. If it matches, advances; otherwise adds error.
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// expectCur checks if the current token matches the expected type.
func (p *Parser) expectCur(t TokenType) bool {
	if p.curToken.Type == t {
		return true
	}
	p.curError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	p.addError("expected next token to be %s, got %s instead", t, p.peekToken.Type)
}

func (p *Parser) curError(t TokenType) {
	p.addError("expected current token to be %s, got %s instead", t, p.curToken.Type)
}

// --- Replace old placeholder methods ---

// Parse is the main entry point called by the service.
// It initializes the lexer and a stateful parser instance for this specific run,
// then calls ParseExpression on the stateful instance.
func (p *Parser) Parse(latexString string) (internalast.Expr, error) {
	l := NewLexer(latexString)
	statefulParser := newStatefulParser(l) // Create a new stateful parser instance
	expr, err := statefulParser.ParseExpression()
	if err != nil {
		// Combine parsing errors if any were collected
		if len(statefulParser.errors) > 0 {
			return nil, fmt.Errorf("parsing failed:\n\t%s", strings.Join(statefulParser.errors, "\n\t"))
		}
		// Otherwise, return the direct error from ParseExpression
		return nil, err
	}
	// Check for errors even if ParseExpression didn't return one directly
	if len(statefulParser.errors) > 0 {
		return nil, fmt.Errorf("parsing failed:\n\t%s", strings.Join(statefulParser.errors, "\n\t"))
	}
	return expr, nil
}

// translate is no longer needed with a custom parser. Remove it.
/*
func (p *Parser) translate(node interface{}) (internalast.Node, error) {
    // ... old placeholder code ...
}
*/
