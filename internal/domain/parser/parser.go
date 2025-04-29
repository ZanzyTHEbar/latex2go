package parser

import (
	"fmt"
	"strconv"
	"strings"

	internalast "github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
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
	LPAREN:   CALL,
	COMMAND:  CALL,
}

// --- Parser Implementation ---

type (
	prefixParseFn func() (internalast.Expr, error)
	infixParseFn  func(internalast.Expr) (internalast.Expr, error)
)

type Parser struct {
	l      *Lexer
	errors []string

	curToken  Token
	peekToken Token

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

func NewParser() *Parser {
	return &Parser{}
}

func newStatefulParser(l *Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[TokenType]prefixParseFn),
		infixParseFns:  make(map[TokenType]infixParseFn),
	}

	p.registerPrefix(IDENT, p.parseIdentifier)
	p.registerPrefix(NUMBER, p.parseNumberLiteral)
	p.registerPrefix(LPAREN, p.parseGroupedExpression)
	p.registerPrefix(MINUS, p.parsePrefixExpression)
	p.registerPrefix(COMMAND, p.parseCommandExpression)

	p.registerInfix(PLUS, p.parseInfixExpression)
	p.registerInfix(MINUS, p.parseInfixExpression)
	p.registerInfix(ASTERISK, p.parseInfixExpression)
	p.registerInfix(SLASH, p.parseInfixExpression)
	p.registerInfix(CARET, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

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

func (p *Parser) ParseExpression() (internalast.Expr, error) {
	expr, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parsing failed:\n\t%s", strings.Join(p.errors, "\n\t"))
	}
	if p.peekToken.Type != EOF {
		p.peekError(EOF) // Expected EOF, got something else
		if len(p.errors) > 0 {
			return nil, fmt.Errorf("parsing failed:\n\t%s", strings.Join(p.errors, "\n\t"))
		}
		return nil, fmt.Errorf("unexpected token '%s' after expression", p.peekToken.Literal)
	}
	return expr, nil
}

func (p *Parser) parseExpression(precedence int) (internalast.Expr, error) {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		err := fmt.Errorf("no prefix parse function found for token %s ('%s')", p.curToken.Type, p.curToken.Literal)
		p.addError("%s", err.Error())
		return nil, err
	}
	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}
	for p.peekToken.Type != EOF && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp, nil
		}
		p.nextToken()
		leftExp, err = infix(leftExp)
		if err != nil {
			return nil, err
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
	return &internalast.Variable{Name: p.curToken.Literal}, nil
}

func (p *Parser) parseNumberLiteral() (internalast.Expr, error) {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		err = fmt.Errorf("could not parse '%s' as float: %w", p.curToken.Literal, err)
		p.addError("%s", err.Error())
		return nil, err
	}
	return &internalast.NumberLiteral{Value: val}, nil
}

func (p *Parser) parsePrefixExpression() (internalast.Expr, error) {
	if p.curToken.Type != MINUS {
		err := fmt.Errorf("expected prefix operator (e.g., '-'), got %s", p.curToken.Type)
		p.addError("%s", err.Error())
		return nil, err
	}
	p.nextToken()
	rightExpr, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}
	return &internalast.BinaryExpr{
		Op:    "*",
		Left:  &internalast.NumberLiteral{Value: -1.0},
		Right: rightExpr,
	}, nil
}

func (p *Parser) parseInfixExpression(left internalast.Expr) (internalast.Expr, error) {
	expr := &internalast.BinaryExpr{
		Op:   p.curToken.Literal,
		Left: left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	var err error
	
	// Special handling for ^ operator to make it right-associative
	if expr.Op == "^" {
		// Pass precedence-1 to give right-side expressions higher precedence
		expr.Right, err = p.parseExpression(precedence - 1)
	} else {
		expr.Right, err = p.parseExpression(precedence)
	}
	
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parseGroupedExpression() (internalast.Expr, error) {
	p.nextToken()
	expr, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if !p.expectPeek(RPAREN) {
		return nil, fmt.Errorf("missing closing parenthesis")
	}
	return expr, nil
}

// --- Enhanced parseCommandExpression for \sum and \prod ---
func (p *Parser) parseCommandExpression() (internalast.Expr, error) {
	funcName := p.curToken.Literal

	// Special handling for \sum and \prod
	if funcName == "sum" || funcName == "prod" {
		isProduct := funcName == "prod"

		// Expect subscript (lower bound): _{i=1}
		if p.peekToken.Type != UNDERSCORE {
			p.addError("expected '_' for lower bound after \\%s", funcName)
			return nil, fmt.Errorf("expected '_' for lower bound after \\%s", funcName)
		}
		p.nextToken() // consume '_'

		if p.peekToken.Type != LBRACE {
			p.addError("expected '{' after '_' in \\%s", funcName)
			return nil, fmt.Errorf("expected '{' after '_' in \\%s", funcName)
		}
		p.nextToken() // consume '{'

		p.nextToken() // move to variable
		varName := ""
		if p.curToken.Type == IDENT {
			varName = p.curToken.Literal
		} else {
			p.addError("expected identifier for summation variable in \\%s", funcName)
			return nil, fmt.Errorf("expected identifier for summation variable in \\%s", funcName)
		}
		p.nextToken() // move to '='
		if p.curToken.Type != EQUALS {
			p.addError("expected '=' after variable in \\%s lower bound", funcName)
			return nil, fmt.Errorf("expected '=' after variable in \\%s lower bound", funcName)
		}
		p.nextToken() // move to lower bound expr
		lower, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		// After parsing the lower bound, expect to see RBRACE as the next token
		if p.peekToken.Type != RBRACE {
			p.addError("expected '}' after lower bound in \\%s", funcName)
			return nil, fmt.Errorf("expected '}' after lower bound in \\%s", funcName)
		}
		p.nextToken() // consume RBRACE

		// Expect superscript (upper bound): ^{n}
		if p.peekToken.Type != CARET {
			p.addError("expected '^' for upper bound after lower bound in \\%s", funcName)
			return nil, fmt.Errorf("expected '^' for upper bound after lower bound in \\%s", funcName)
		}
		p.nextToken() // consume '}'
		p.nextToken() // consume '^'
		if p.curToken.Type != LBRACE {
			p.addError("expected '{' after '^' in \\%s", funcName)
			return nil, fmt.Errorf("expected '{' after '^' in \\%s", funcName)
		}
		p.nextToken() // move to upper bound expr
		upper, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		// After parsing the upper bound, expect to see RBRACE as the next token
		if p.peekToken.Type != RBRACE {
			p.addError("expected '}' after upper bound in \\%s", funcName)
			return nil, fmt.Errorf("expected '}' after upper bound in \\%s", funcName)
		}
		p.nextToken() // consume RBRACE
		p.nextToken() // advance to body token

		body, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}

		return &internalast.SumExpr{
			IsProduct: isProduct,
			Var:       varName,
			Lower:     lower,
			Upper:     upper,
			Body:      body,
		}, nil
	}

	args := []internalast.Expr{}
	for p.peekToken.Type == LBRACE {
		p.nextToken() // consume LBRACE
		// Check for empty argument: {} immediately after command
		if p.peekToken.Type == RBRACE {
			err := fmt.Errorf("argument expression cannot be empty inside {} for command \\%s", funcName)
			p.addError("%s", err.Error())
			return nil, err
		}
		p.nextToken() // consume token after LBRACE (start of expression)
		argExpr, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, argExpr)
		if p.peekToken.Type != RBRACE {
			if p.peekToken.Type == EOF {
				// Use a more specific error message for EOF
				err := fmt.Errorf("missing '}' after argument for command \\%s", funcName)
				p.addError("%s", err.Error())
				return nil, err
			}
			p.peekError(RBRACE)
			return nil, fmt.Errorf("missing '}' after argument for command \\%s", funcName)
		}
		p.nextToken() // consume RBRACE
	}

	if len(args) == 0 && funcName != "sum" && funcName != "prod" { // Allow sum/prod to have no {} args initially
		err := fmt.Errorf("expected '{' arguments after command '\\%s', got %s", funcName, p.peekToken.Type)
		p.addError("%s", err.Error())
		return nil, err
	}

	// Check if we have the required number of arguments
	requiredArgs := -1
	switch strings.ToLower(funcName) {
	case "frac":
		requiredArgs = 2
	case "sqrt", "sin", "cos", "tan":
		requiredArgs = 1
	}

	if requiredArgs != -1 && len(args) != requiredArgs {
		err := fmt.Errorf("\\%s requires %d argument(s), got %d", funcName, requiredArgs, len(args))
		p.addError("%s", err.Error())
		return nil, err
	}

	// Check for unexpected tokens after the function and its arguments
	if p.peekToken.Type != EOF && p.peekToken.Type != RPAREN && 
	   !(p.peekToken.Type == PLUS || p.peekToken.Type == MINUS || 
	     p.peekToken.Type == ASTERISK || p.peekToken.Type == SLASH || 
	     p.peekToken.Type == CARET) {
		err := fmt.Errorf("unexpected token '%s' after expression", p.peekToken.Type)
		p.addError("%s", err.Error())
		return nil, err
	}

	return &internalast.FuncCall{
		FuncName: funcName,
		Args:     args,
	}, nil
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	p.addError("expected next token to be %s, got %s ('%s') instead", t, p.peekToken.Type, p.peekToken.Literal)
}

func (p *Parser) Parse(latexString string) (internalast.Expr, error) {
	l := NewLexer(latexString)
	statefulParser := newStatefulParser(l)
	expr, err := statefulParser.ParseExpression()
	if err != nil {
		if len(statefulParser.errors) > 0 {
			return nil, fmt.Errorf("parsing failed:\n\t%s", strings.Join(statefulParser.errors, "\n\t"))
		}
		return nil, err
	}
	if len(statefulParser.errors) > 0 {
		return nil, fmt.Errorf("parsing failed:\n\t%s", strings.Join(statefulParser.errors, "\n\t"))
	}
	return expr, nil
}
