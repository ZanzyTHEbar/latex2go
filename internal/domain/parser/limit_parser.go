package parser

import (
	"fmt"
	"strings"

	internalast "github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
)

// parseLimitExpression handles parsing of limit expressions like:
// \lim_{x \to 0} or \lim{x \to 0}
func (p *Parser) parseLimitExpression(braceStyle bool) (internalast.Expr, error) {
	var varName string

	// If we're not in brace style, then we expect underscore followed by a brace
	if !braceStyle {
		// Check for opening brace after underscore
		if p.peekToken.Type != LBRACE {
			p.addError("expected '{' after '_' in \\lim")
			return nil, fmt.Errorf("expected '{' after '_' in \\lim")
		}
		p.nextToken() // consume '{'
	}

	// Next token should be the variable
	p.nextToken()
	if p.curToken.Type != IDENT {
		p.addError("expected identifier for limit variable")
		return nil, fmt.Errorf("expected identifier for limit variable in \\lim")
	}

	varName = p.curToken.Literal
	p.nextToken() // Move past variable name

	// Now handle different variations of "to" notation
	// First, skip any non-significant tokens until we find something interesting
	for p.curToken.Type != IDENT && p.curToken.Type != COMMAND &&
		p.curToken.Type != NUMBER && p.curToken.Type != RBRACE {
		p.nextToken()
	}

	// Now check for "to" token in various forms
	toFound := false

	// Try different patterns for finding "to"
	if p.curToken.Type == IDENT && strings.ToLower(p.curToken.Literal) == "to" {
		// Case 1: "to" as an identifier
		toFound = true
		p.nextToken() // Move past "to"
	} else if p.curToken.Type == COMMAND && p.curToken.Literal == "to" {
		// Case 2: "\to" as a LaTeX command
		toFound = true
		p.nextToken() // Move past "\to"
	} else if p.curToken.Type == COMMAND && p.peekToken.Type == IDENT &&
		(p.peekToken.Literal == "o" || p.peekToken.Literal == "to") {
		// Case 3: "\t" followed by "o" (tokenized separately)
		toFound = true
		p.nextToken() // Move to "o" or "to" token
		p.nextToken() // Move past it
	} else if p.curToken.Type == IDENT && p.curToken.Literal == "t" &&
		p.peekToken.Type == IDENT && p.peekToken.Literal == "o" {
		// Case 4: "t" and "o" tokenized separately
		toFound = true
		p.nextToken() // Move to "o"
		p.nextToken() // Move past "o"
	}

	// If we couldn't find a "to" token after several attempts,
	// just assume it's implied and continue (more resilient)
	if !toFound {
		// Log the situation but don't fail the parse
		p.addError("warning: couldn't find 'to' in limit expression, assuming implied")
	}

	// Skip any additional whitespace or non-significant tokens
	for p.curToken.Type != IDENT && p.curToken.Type != NUMBER &&
		p.curToken.Type != COMMAND && p.curToken.Type != RBRACE {
		p.nextToken()
	}

	// Now parse the approach value
	approaches, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	// Check for closing brace
	if p.peekToken.Type != RBRACE {
		p.addError("expected '}' after approach value in \\lim")
		return nil, fmt.Errorf("expected '}' after approach value in \\lim")
	}
	// Consume closing brace
	p.nextToken() // Consume the closing brace '}'

	// Move past the closing brace to get ready for the body expression
	// This is the key fix - ensuring we're positioned correctly for parsing the body
	p.nextToken()

	// Now parse the body expression
	body, err := p.parseExpression(LOWEST)
	if err != nil {
		p.addError("failed to parse limit body expression: %s", err)
		return nil, fmt.Errorf("failed to parse limit body expression: %w", err)
	}

	return &internalast.LimitExpr{
		Var:        varName,
		Approaches: approaches,
		Body:       body,
	}, nil
}
