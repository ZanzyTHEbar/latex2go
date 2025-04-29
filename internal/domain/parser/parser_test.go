package parser

import (
	"fmt"
	"strings"
	"testing"

	internalast "github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to check parser errors
func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("Parser has %d errors:", len(errors))
	for _, msg := range errors {
		t.Errorf(" - %s", msg)
	}
	t.FailNow()
}

// Helper to test number literals
func testNumberLiteral(t *testing.T, expr internalast.Expr, expected float64) bool {
	t.Helper()
	num, ok := expr.(*internalast.NumberLiteral)
	if !ok {
		t.Errorf("expr not *ast.NumberLiteral. got=%T", expr)
		return false
	}
	if num.Value != expected {
		t.Errorf("num.Value not %f. got=%f", expected, num.Value)
		return false
	}
	return true
}

// Helper to test variable identifiers
func testVariable(t *testing.T, expr internalast.Expr, expected string) bool {
	t.Helper()
	ident, ok := expr.(*internalast.Variable)
	if !ok {
		t.Errorf("expr not *ast.Variable. got=%T", expr)
		return false
	}
	if ident.Name != expected {
		t.Errorf("ident.Name not %s. got=%s", expected, ident.Name)
		return false
	}
	return true
}

// Helper to test binary expressions
func testBinaryExpr(t *testing.T, expr internalast.Expr, expectedLeft interface{}, expectedOp string, expectedRight interface{}) bool {
	t.Helper()
	binExpr, ok := expr.(*internalast.BinaryExpr)
	if !ok {
		t.Errorf("expr is not ast.BinaryExpr. got=%T(%s)", expr, expr)
		return false
	}

	if !testLiteralExpression(t, binExpr.Left, expectedLeft) {
		return false
	}

	if binExpr.Op != expectedOp {
		t.Errorf("expr.Op is not '%s'. got=%q", expectedOp, binExpr.Op)
		return false
	}

	if !testLiteralExpression(t, binExpr.Right, expectedRight) {
		return false
	}

	return true
}

// Helper to test literal expressions (number or variable)
func testLiteralExpression(t *testing.T, expr internalast.Expr, expected interface{}) bool {
	t.Helper()
	switch v := expected.(type) {
	case int:
		return testNumberLiteral(t, expr, float64(v))
	case int64:
		return testNumberLiteral(t, expr, float64(v))
	case float64:
		return testNumberLiteral(t, expr, v)
	case string:
		return testVariable(t, expr, v)
	default:
		t.Errorf("type of expr not handled. got=%T", expr)
		return false
	}
}

func TestParser_BasicArithmetic(t *testing.T) {
	tests := []struct {
		input         string
		expectedLeft  interface{}
		expectedOp    string
		expectedRight interface{}
	}{
		{"a + b", "a", "+", "b"},
		{"x - 5", "x", "-", 5.0},
		{"y * 3.14", "y", "*", 3.14},
		{"10 / z", 10.0, "/", "z"},
		{"a + b * c", "a", "+", nil}, // Tests precedence (b*c is right node)
		{"(a + b) * c", nil, "*", "c"}, // Tests grouping
		{"2 * (x - y)", 2.0, "*", nil}, // Tests grouping
		{"a / b + c", nil, "+", "c"}, // Tests precedence (a/b is left node)
		{"a - b / c", "a", "-", nil}, // Tests precedence (b/c is right node)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := newStatefulParser(l)
			expr, err := p.ParseExpression()
			require.NoError(t, err)
			checkParserErrors(t, p)
			require.NotNil(t, expr)

			binExpr, ok := expr.(*internalast.BinaryExpr)
			require.True(t, ok, "Expected BinaryExpr")

			// Simplified checks for precedence tests
			if tt.expectedLeft == nil {
				// Check the operator of the top-level expression
				assert.Equal(t, tt.expectedOp, binExpr.Op)
				// Check the right operand if specified
				if tt.expectedRight != nil {
					testLiteralExpression(t, binExpr.Right, tt.expectedRight)
				}
				// Further checks could inspect the nested structure (e.g., binExpr.Left)
			} else if tt.expectedRight == nil {
				// Check the operator of the top-level expression
				assert.Equal(t, tt.expectedOp, binExpr.Op)
				// Check the left operand if specified
				if tt.expectedLeft != nil {
					testLiteralExpression(t, binExpr.Left, tt.expectedLeft)
				}
				// Further checks could inspect the nested structure (e.g., binExpr.Right)
			} else {
				// Standard check for simple binary expressions
				testBinaryExpr(t, expr, tt.expectedLeft, tt.expectedOp, tt.expectedRight)
			}
		})
	}
}

func TestParser_Exponentiation(t *testing.T) {
	tests := []struct {
		input         string
		expectedLeft  interface{}
		expectedOp    string
		expectedRight interface{}
	}{
		{"a ^ b", "a", "^", "b"},
		{"x ^ 2", "x", "^", 2.0},
		{"3 ^ y", 3.0, "^", "y"},
		{"a ^ b ^ c", "a", "^", nil}, // Right-associative (a ^ (b^c))
		{"(a ^ b) ^ c", nil, "^", "c"}, // Grouping overrides associativity
		{"a * b ^ c", "a", "*", nil}, // Precedence: ^ higher than *
		{"a ^ b * c", nil, "*", "c"}, // Precedence: ^ higher than *
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := newStatefulParser(l)
			expr, err := p.ParseExpression()
			require.NoError(t, err)
			checkParserErrors(t, p)
			require.NotNil(t, expr)

			binExpr, ok := expr.(*internalast.BinaryExpr)
			require.True(t, ok, "Expected BinaryExpr")

			// Simplified checks for precedence/associativity tests
			if tt.expectedLeft == nil {
				assert.Equal(t, tt.expectedOp, binExpr.Op)
				if tt.expectedRight != nil {
					testLiteralExpression(t, binExpr.Right, tt.expectedRight)
				}
			} else if tt.expectedRight == nil {
				assert.Equal(t, tt.expectedOp, binExpr.Op)
				if tt.expectedLeft != nil {
					testLiteralExpression(t, binExpr.Left, tt.expectedLeft)
					}
				
				// For the case a ^ b ^ c, test that the right side is actually (b ^ c)
				if tt.input == "a ^ b ^ c" {
					rightBin, ok := binExpr.Right.(*internalast.BinaryExpr)
					require.True(t, ok, "Expected Right of 'a ^ b ^ c' to be BinaryExpr due to right-associativity")
					assert.Equal(t, "^", rightBin.Op, "Expected '^' operator in nested right expression")
					testLiteralExpression(t, rightBin.Left, "b")
					testLiteralExpression(t, rightBin.Right, "c")
				}
			} else {
				testBinaryExpr(t, expr, tt.expectedLeft, tt.expectedOp, tt.expectedRight)
			}
		})
	}
}

func TestParser_UnaryMinus(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"-a", "a"},
		{"-5", 5.0},
		{"- (a + b)", nil}, // Check negation of a group
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := newStatefulParser(l)
			expr, err := p.ParseExpression()
			require.NoError(t, err)
			checkParserErrors(t, p)
			require.NotNil(t, expr)

			// Unary minus is parsed as BinaryExpr{Left: -1.0, Op: "*", Right: <operand>}
			binExpr, ok := expr.(*internalast.BinaryExpr)
			require.True(t, ok, "Expected BinaryExpr for unary minus representation")
			assert.True(t, testNumberLiteral(t, binExpr.Left, -1.0), "Left operand should be -1.0")
			assert.Equal(t, "*", binExpr.Op, "Operator should be *")

			if tt.expectedValue != nil {
				testLiteralExpression(t, binExpr.Right, tt.expectedValue)
			} else {
				// Check if the right side is a non-literal expression (like another BinaryExpr for grouped negation)
				_, isNum := binExpr.Right.(*internalast.NumberLiteral)
				_, isVar := binExpr.Right.(*internalast.Variable)
				assert.False(t, isNum || isVar, "Right operand should be a complex expression for grouped negation, not a literal")
			}
		})
	}
}

func TestParser_FunctionCalls(t *testing.T) {
	tests := []struct {
		input          string
		expectedFunc   string
		expectedArgs   []interface{}
		expectErrorMsg string // If non-empty, expect an error containing this msg
	}{
		{`\sqrt{x}`, "sqrt", []interface{}{"x"}, ""},
		{`\sin{y}`, "sin", []interface{}{"y"}, ""},
		{`\cos{3.14}`, "cos", []interface{}{3.14}, ""},
		{`\tan{(a+b)}`, "tan", []interface{}{nil}, ""}, // Argument is a complex expression
		{`\frac{a}{b}`, "frac", []interface{}{"a", "b"}, ""},
		{`\frac{1}{x+y}`, "frac", []interface{}{1.0, nil}, ""},
		{`\sqrt{x+y}`, "sqrt", []interface{}{nil}, ""},
		// Error cases
		{`\sqrt`, "sqrt", nil, "expected '{' arguments after command"},
		{`\sqrt{}`, "sqrt", nil, "argument expression cannot be empty"}, 
		{`\sqrt{x`, "sqrt", nil, "missing '}'"},  // Just check for any error containing missing '}'
		{`\frac{a}`, "frac", nil, "requires 2 argument(s), got 1"},
		{`\frac{}{b}`, "frac", nil, "argument expression cannot be empty"}, 
		{`\frac{a}{}`, "frac", nil, "argument expression cannot be empty"}, 
		{`\frac{a}{b}{c}`, "frac", nil, "\\frac requires 2 argument(s), got 3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := newStatefulParser(l)
			expr, err := p.ParseExpression()

			if tt.expectErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrorMsg)
				// Also check parser errors list if populated
				errors := p.Errors()
				if len(errors) > 0 {
					assert.Contains(t, errors[len(errors)-1], tt.expectErrorMsg)
				}
				return // Don't check expression if error expected
			}

			require.NoError(t, err)
			checkParserErrors(t, p)
			require.NotNil(t, expr)

			callExpr, ok := expr.(*internalast.FuncCall)
			require.True(t, ok, "Expected FuncCall")
			assert.Equal(t, tt.expectedFunc, callExpr.FuncName)
			require.Len(t, callExpr.Args, len(tt.expectedArgs))

			for i, expectedArg := range tt.expectedArgs {
				if expectedArg == nil {
					// Check that the argument is a non-literal expression
					_, isNum := callExpr.Args[i].(*internalast.NumberLiteral)
					_, isVar := callExpr.Args[i].(*internalast.Variable)
					assert.False(t, isNum || isVar, fmt.Sprintf("Arg %d should be a complex expression, not a literal", i))
				} else {
					testLiteralExpression(t, callExpr.Args[i], expectedArg)
				}
			}
		})
	}
}

func TestParser_Errors(t *testing.T) {
	tests := []struct {
		input          string
		expectErrorMsg string
	}{
		{`1 +`, "no prefix parse function found for token EOF"},
		{`* 2`, "no prefix parse function found for token ASTERISK"},
		{`( a + b`, "missing closing parenthesis"},
		{`a + b )`, "expected next token to be EOF, got RPAREN"},
		{`\sqrt{x} y`, "unexpected token 'IDENT' after expression"}, // Update to match actual error
		{`1.2.3`, "expected next token to be EOF, got ILLEGAL"},
		{`{`, "no prefix parse function found for token LBRACE"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := newStatefulParser(l)
			_, err := p.ParseExpression()
			require.Error(t, err)
			assert.NotEmpty(t, p.Errors())
			// Check if the final error message or one of the parser errors contains the expected substring
			found := false
			if strings.Contains(err.Error(), tt.expectErrorMsg) { // Use strings.Contains for safer check
				found = true
			}
			for _, pErr := range p.Errors() {
				if strings.Contains(pErr, tt.expectErrorMsg) { // Use strings.Contains for safer check
					found = true
					break
				}
			}
			assert.True(t, found, fmt.Sprintf("Expected error message substring '%s' not found in final error ('%s') or parser errors list ('%v')", tt.expectErrorMsg, err.Error(), p.Errors()))
		})
	}
}
