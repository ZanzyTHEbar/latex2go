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
	PREFIX   // -X (unary minus)
	POSTFIX  // X! (factorial)
	CALL     // myFunction(X) or \command{X}
)

var precedences = map[TokenType]int{
	PLUS:       SUM,
	MINUS:      SUM,
	ASTERISK:   PRODUCT,
	SLASH:      PRODUCT,
	CARET:      EXPONENT,
	EXCLAMATION: POSTFIX, // Factorial has higher precedence
	LPAREN:     CALL,
	COMMAND:    CALL,
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
	p.registerPrefix(BEGIN, p.parsePiecewiseExpression) // Add parsing for \begin{cases}

	p.registerInfix(PLUS, p.parseInfixExpression)
	p.registerInfix(MINUS, p.parseInfixExpression)
	p.registerInfix(ASTERISK, p.parseInfixExpression)
	p.registerInfix(SLASH, p.parseInfixExpression)
	p.registerInfix(CARET, p.parseInfixExpression)
	p.registerInfix(EXCLAMATION, p.parseFactorialExpression) // Add factorial parsing

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

	// Special handling for limit expressions with underscore notation
	if funcName == "lim" {
		if p.peekToken.Type == UNDERSCORE {
			// Handle \lim_{x \to a} notation directly
			p.nextToken() // consume underscore
			return p.parseLimitExpression(false)
		} else {
			// Handle \lim without underscore - maybe it's using plain text 
			// like \lim x \to 0 or will have arguments in braces later
			// For now, just pass it to the standard argument handling
		}
	}

	// Special handling for \sum and \prod
	if (funcName == "sum" || funcName == "prod") {
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
	
	// Special handling for \int (integral)
	if funcName == "int" {
		isDefinite := false
		var lower, upper internalast.Expr
		
		// Check if we have a definite integral with bounds
		if p.peekToken.Type == UNDERSCORE {
			isDefinite = true
			
			// Parse lower bound: _{a}
			p.nextToken() // consume '_'
			if p.peekToken.Type != LBRACE {
				p.addError("expected '{' after '_' in \\%s", funcName)
				return nil, fmt.Errorf("expected '{' after '_' in \\%s", funcName)
			}
			p.nextToken() // consume '{'
			p.nextToken() // move to lower bound expression
			var err error
			lower, err = p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}
			
			// After parsing the lower bound, expect to see RBRACE
			if p.peekToken.Type != RBRACE {
				p.addError("expected '}' after lower bound in \\%s", funcName)
				return nil, fmt.Errorf("expected '}' after lower bound in \\%s", funcName)
			}
			p.nextToken() // consume RBRACE
			
			// Parse upper bound: ^{b}
			if p.peekToken.Type != CARET {
				p.addError("expected '^' for upper bound after lower bound in \\%s", funcName)
				return nil, fmt.Errorf("expected '^' for upper bound after lower bound in \\%s", funcName)
			}
			p.nextToken() // consume '^'
			
			if p.peekToken.Type != LBRACE {
				p.addError("expected '{' after '^' in \\%s", funcName)
				return nil, fmt.Errorf("expected '{' after '^' in \\%s", funcName)
			}
			p.nextToken() // consume '{'
			p.nextToken() // move to upper bound expression
			
			upper, err = p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}
			
			// After parsing the upper bound, expect to see RBRACE
			if p.peekToken.Type != RBRACE {
				p.addError("expected '}' after upper bound in \\%s", funcName)
				return nil, fmt.Errorf("expected '}' after upper bound in \\%s", funcName)
			}
			p.nextToken() // consume RBRACE
		}
		
		// Parse the body of the integral
		p.nextToken() // Move to the body expression
		body, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		
		// Find the differential variable (e.g., "dx" in \int f(x) dx)
		// Look for a command or identifier that should represent the differential
		var integrationVar string
		if p.peekToken.Type == IDENT && strings.HasPrefix(p.peekToken.Literal, "d") {
			// Extract the variable name from "dx", "dy", etc.
			integrationVar = strings.TrimPrefix(p.peekToken.Literal, "d")
			p.nextToken() // consume the differential
		} else {
			// If no differential is specified, default to "x"
			integrationVar = "x"
		}
		
		return &internalast.IntegralExpr{
			IsDefinite: isDefinite,
			Var:        integrationVar,
			Lower:      lower,
			Upper:      upper,
			Body:       body,
		}, nil
	}

	args := []internalast.Expr{}
	
	// Special handling for \lim
	if funcName == "lim" && p.peekToken.Type == LBRACE {
		p.nextToken() // consume LBRACE
		
		// Read the raw content of the argument to handle "x \to 0" format
		if p.peekToken.Type == IDENT {
			varName := p.peekToken.Literal
			p.nextToken() // move past variable
			
			// Check if next tokens form "to 0" pattern
			if p.peekToken.Type == IDENT && p.peekToken.Literal == "to" {
				p.nextToken() // consume "to"
				p.nextToken() // move to the number
				
				// Parse the approach value
				if p.curToken.Type == NUMBER {
					approachVal, _ := strconv.ParseFloat(p.curToken.Literal, 64)
					approaches := &internalast.NumberLiteral{Value: approachVal}
					
					// Consume RBRACE
					if p.peekToken.Type == RBRACE {
						p.nextToken() // consume RBRACE
						p.nextToken() // move to next token for body
						
						// Parse the body expression
						body, err := p.parseExpression(LOWEST)
						if err != nil {
							return nil, err
						}
						
						return &internalast.LimitExpr{
							Var:        varName,
							Approaches: approaches,
							Body:       body,
						}, nil
					}
				}
			}
		}
		
		// If we couldn't parse as a limit expression, rewind and parse normally
		// This is just a partial implementation - a real one would need to rewind properly
	}
	
	// Standard argument parsing
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
		// Special case for derivatives: \frac{d}{dx} or \frac{\partial}{\partial x}
		if len(args) == 2 {
			// Check if this might be a derivative
			if numExpr, ok := args[0].(*internalast.Variable); ok {
				if numExpr.Name == "d" || numExpr.Name == "\\partial" {
					// Check denominator for dx or \partial x pattern
					if denExpr, ok := args[1].(*internalast.Variable); ok {
						if strings.HasPrefix(denExpr.Name, "d") || strings.HasPrefix(denExpr.Name, "\\partial ") {
							// This appears to be a derivative setup
							// Extract the variable of differentiation
							var diffVar string
							var isPartial bool
							
							if strings.HasPrefix(denExpr.Name, "d") {
								diffVar = strings.TrimPrefix(denExpr.Name, "d")
								isPartial = false
							} else {
								diffVar = strings.TrimPrefix(denExpr.Name, "\\partial ")
								isPartial = true
							}
							
							// Look ahead to capture the expression being differentiated
							if p.peekToken.Type == IDENT || p.peekToken.Type == COMMAND || 
							   p.peekToken.Type == LPAREN || p.peekToken.Type == NUMBER {
								body, err := p.parseExpression(LOWEST)
								if err != nil {
									return nil, err
								}
								
								return &internalast.DerivativeExpr{
									IsPartial: isPartial,
									Var:       diffVar,
									Order:     1, // First-order derivative
									Body:      body,
								}, nil
							}
						}
					}
				}
			}
		}
		requiredArgs = 2	
	case "lim":
		// We now handle the underscore notation directly in parseCommandExpression
		// Here we just handle braced notation and direct variable notation
		
		// Skip any whitespace or non-brace tokens to find either a brace or the variable directly
		maxLookahead := 5 // Maximum number of tokens to look ahead
		for i := 1; i <= maxLookahead; i++ {
			peekType, peekLit := p.peekNTokens(i)
			
			// If we find an opening brace, navigate to it and parse the limit
			if peekType == LBRACE {
				// Skip to the brace
				for j := 1; j <= i; j++ {
					p.nextToken()
				}
				return p.parseLimitExpression(true)
			}
			
			// If we find an identifier (possibly the limit variable directly), 
			// navigate to it and try to parse as a limit
			if peekType == IDENT && peekLit != "" && peekLit != "to" {
				// Skip to the identifier
				for j := 1; j < i; j++ {
					p.nextToken()
				}
				
				// Create a synthetic environment as if we had braces
				varName := peekLit
				
				// Skip the variable
				p.nextToken()
				
				// Look for "to" token
				for k := 0; k < 3; k++ { // Try up to 3 tokens ahead for "to"
					if p.curToken.Type == IDENT && p.curToken.Literal == "to" ||
					   (p.curToken.Type == COMMAND && p.curToken.Literal == "to") {
						p.nextToken() // Skip "to"
						break
					}
					p.nextToken()
				}
				
				// Parse approach value
				approaches, err := p.parseExpression(LOWEST)
				if err != nil {
					return nil, err
				}
				
				// Parse body expression
				body, err := p.parseExpression(LOWEST)
				if err != nil {
					return nil, err
				}
				
				return &internalast.LimitExpr{
					Var:        varName,
					Approaches: approaches,
					Body:       body,
				}, nil
			}
		}
		
		// If we didn't find a limit pattern, fall back to regular function parsing
		requiredArgs = 1
	case "sqrt", "sin", "cos", "tan":
		requiredArgs = 1
	}

	if requiredArgs != -1 && len(args) != requiredArgs {
		err := fmt.Errorf("\\%s requires %d argument(s), got %d", funcName, requiredArgs, len(args))
		p.addError("%s", err.Error())
		return nil, err
	}

	// Check for unexpected tokens after the function and its arguments
	// Valid tokens after a function expression can be:
	// - EOF (end of input)
	// - RPAREN (closing parenthesis for grouped expressions)
	// - RBRACE (closing brace for nested LaTeX commands)
	// - Operators (PLUS, MINUS, ASTERISK, SLASH, CARET)
	if p.peekToken.Type != EOF && p.peekToken.Type != RPAREN && p.peekToken.Type != RBRACE && 
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

func (p *Parser) parsePiecewiseExpression() (internalast.Expr, error) {
	// Check if this is a \begin{cases} environment
	if p.curToken.Type != BEGIN {
		return nil, fmt.Errorf("expected \\begin for piecewise expression")
	}
	
	// Check for the opening brace and "cases" environment
	if p.peekToken.Type != LBRACE {
		p.addError("expected '{' after \\begin for cases environment")
		return nil, fmt.Errorf("expected '{' after \\begin for cases environment")
	}
	p.nextToken() // consume '{'
	
	// Read the environment type (should be "cases")
	p.nextToken() // move to environment identifier
	if p.curToken.Type != IDENT || p.curToken.Literal != "cases" {
		p.addError("expected 'cases' for piecewise environment")
		return nil, fmt.Errorf("expected 'cases' for piecewise environment")
	}
	
	// Check for closing brace
	if p.peekToken.Type != RBRACE {
		p.addError("expected '}' after 'cases' in \\begin")
		return nil, fmt.Errorf("expected '}' after 'cases' in \\begin")
	}
	p.nextToken() // consume '}'
	p.nextToken() // move past '}'
	
	// Now parse the cases until we reach \end{cases}
	cases := []internalast.PiecewiseCase{}
	
	for p.curToken.Type != END {
		// Parse the case value (expression)
		value, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		
		// Check for the condition separator (usually &)
		// Note: This is a simplification, as LaTeX typically uses & for alignment
		var condition internalast.Expr
		if p.peekToken.Type == IDENT && p.peekToken.Literal == "&" {
			p.nextToken() // consume the alignment marker
			
			// Parse the condition expression
			condition, err = p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}
		}
		
		// Add the case
		cases = append(cases, internalast.PiecewiseCase{
			Value:     value,
			Condition: condition,
		})
		
		// Look for case separator (usually \\)
		// Again, this is a simplification
		if p.peekToken.Type == COMMAND && p.peekToken.Literal == "\\" {
			p.nextToken() // consume the line break
		}
		
		// Move to the next token to continue parsing
		p.nextToken()
	}
	
	// Now we should be at \end{cases}
	if p.curToken.Type != END {
		p.addError("expected \\end for cases environment")
		return nil, fmt.Errorf("expected \\end for cases environment")
	}
	
	// Check for the closing environment tag
	if p.peekToken.Type != LBRACE {
		p.addError("expected '{' after \\end")
		return nil, fmt.Errorf("expected '{' after \\end")
	}
	p.nextToken() // consume '{'
	
	// Check that we're closing the "cases" environment
	p.nextToken() // move to environment identifier
	if p.curToken.Type != IDENT || p.curToken.Literal != "cases" {
		p.addError("expected 'cases' in \\end{}")
		return nil, fmt.Errorf("expected 'cases' in \\end{}")
	}
	
	// Check for closing brace
	if p.peekToken.Type != RBRACE {
		p.addError("expected '}' after 'cases' in \\end")
		return nil, fmt.Errorf("expected '}' after 'cases' in \\end")
	}
	p.nextToken() // consume '}'
	
	return &internalast.PiecewiseExpr{
		Cases: cases,
	}, nil
}

func (p *Parser) parseFactorialExpression(left internalast.Expr) (internalast.Expr, error) {
	expr := &internalast.FactorialExpr{
		Value: left,
	}
	p.nextToken() // Consume the '!' token
	return expr, nil
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

// peekNTokens peeks ahead n tokens and returns the token type and literal
// Since we can't easily peek ahead, we'll need to create a copy of the lexer state
// and advance it manually
func (p *Parser) peekNTokens(n int) (TokenType, string) {
	if n <= 0 {
		return p.curToken.Type, p.curToken.Literal
	}
	if n == 1 {
		return p.peekToken.Type, p.peekToken.Literal
	}
	
	// Create a temporary copy of the lexer at current position
	// This is a basic implementation that handles enough of the limit expression cases
	curInput := p.l.input
	curPos := p.l.position
	
	// Skip current token and peek token
	skipCount := 2
	
	// Simple character-based forward scan to find the nth non-whitespace token
	for i := curPos; i < len(curInput) && skipCount < n; i++ {
		// Skip whitespace
		if curInput[i] == ' ' || curInput[i] == '\t' || curInput[i] == '\n' || curInput[i] == '\r' {
			continue
		}
		
		// Check if we have a token boundary (simple approximation)
		if curInput[i] == '{' || curInput[i] == '}' || curInput[i] == '(' || curInput[i] == ')' || 
		   curInput[i] == '+' || curInput[i] == '-' || curInput[i] == '*' || curInput[i] == '/' ||
		   curInput[i] == '^' || curInput[i] == '_' || curInput[i] == '\\' {
			skipCount++
			
			// If we've found the nth token, return its type
			if skipCount == n {
				switch curInput[i] {
				case '{':
					return LBRACE, "{"
				case '}':
					return RBRACE, "}"
				case '(':
					return LPAREN, "("
				case ')':
					return RPAREN, ")"
				case '+':
					return PLUS, "+"
				case '-':
					return MINUS, "-"
				case '*':
					return ASTERISK, "*"
				case '/':
					return SLASH, "/"
				case '^':
					return CARET, "^"
				case '_':
					return UNDERSCORE, "_"
				case '\\':
					return COMMAND, "\\"
				default:
					return ILLEGAL, string(curInput[i])
				}
			}
		}
	}
	
	// If we can't peek that far ahead, return EOF
	return EOF, ""
}
