package parser

import (
	"testing"

	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast" // Corrected import path

	"github.com/stretchr/testify/assert" // Import testify/assert
)

func TestParser(t *testing.T) {
	// Placeholder test - more tests should be added here
	t.Run("Simple Addition", func(t *testing.T) {
		input := "a + b"
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)         // Use newStatefulParser
		expr, parseErr := parser.ParseExpression() // Call ParseExpression() without args, capture single error

		// Check collected errors first
		assert.Empty(t, parser.Errors(), "Parser reported errors for valid input '%s': %v", input, parser.Errors())
		// Check the returned error as well
		assert.NoError(t, parseErr, "ParseExpression returned an unexpected error for '%s'", input)
		// Check if expression was parsed
		assert.NotNil(t, expr, "Parser returned nil expression for valid input '%s'", input)

		// Add specific AST assertions
		binExpr, ok := expr.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected a BinaryExpr")
		if ok { // Only proceed if type assertion is successful
			assert.Equal(t, "+", binExpr.Op)
			// Assert left operand (Variable 'a')
			leftVar, okLeft := binExpr.Left.(*ast.Variable)
			assert.True(t, okLeft, "Expected left operand to be Variable")
			if okLeft {
				assert.Equal(t, "a", leftVar.Name)
			}
			// Assert right operand (Variable 'b')
			rightVar, okRight := binExpr.Right.(*ast.Variable)
			assert.True(t, okRight, "Expected right operand to be Variable")
			if okRight {
				assert.Equal(t, "b", rightVar.Name)
			}
		}
	})

	t.Run("Simple Subtraction", func(t *testing.T) {
		input := "a - b"
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)
		expr, parseErr := parser.ParseExpression()

		assert.Empty(t, parser.Errors())
		assert.NoError(t, parseErr)
		assert.NotNil(t, expr)
		binExpr, ok := expr.(*ast.BinaryExpr)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, "-", binExpr.Op)
			assert.IsType(t, &ast.Variable{}, binExpr.Left)
			assert.IsType(t, &ast.Variable{}, binExpr.Right)
			assert.Equal(t, "a", binExpr.Left.(*ast.Variable).Name)
			assert.Equal(t, "b", binExpr.Right.(*ast.Variable).Name)
		}
	})

	t.Run("Multiplication and Division", func(t *testing.T) {
		input := "a * b / c" // Should parse as (a * b) / c
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)
		expr, parseErr := parser.ParseExpression()

		assert.Empty(t, parser.Errors())
		assert.NoError(t, parseErr)
		assert.NotNil(t, expr)

		// Expect: BinaryExpr{Op: "/", Left: BinaryExpr{Op: "*", Left: "a", Right: "b"}, Right: "c"}
		divExpr, ok := expr.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected outer expression to be division")
		if !ok {
			return
		}
		assert.Equal(t, "/", divExpr.Op)

		mulExpr, ok := divExpr.Left.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected left side of division to be multiplication")
		if !ok {
			return
		}
		assert.Equal(t, "*", mulExpr.Op)

		assert.IsType(t, &ast.Variable{}, mulExpr.Left)
		assert.Equal(t, "a", mulExpr.Left.(*ast.Variable).Name)
		assert.IsType(t, &ast.Variable{}, mulExpr.Right)
		assert.Equal(t, "b", mulExpr.Right.(*ast.Variable).Name)

		assert.IsType(t, &ast.Variable{}, divExpr.Right)
		assert.Equal(t, "c", divExpr.Right.(*ast.Variable).Name)
	})

	t.Run("Operator Precedence", func(t *testing.T) {
		input := "a + b * c" // Should parse as a + (b * c)
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)
		expr, parseErr := parser.ParseExpression()

		assert.Empty(t, parser.Errors())
		assert.NoError(t, parseErr)
		assert.NotNil(t, expr)

		// Expect: BinaryExpr{Op: "+", Left: "a", Right: BinaryExpr{Op: "*", Left: "b", Right: "c"}}
		addExpr, ok := expr.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected outer expression to be addition")
		if !ok {
			return
		}
		assert.Equal(t, "+", addExpr.Op)

		assert.IsType(t, &ast.Variable{}, addExpr.Left)
		assert.Equal(t, "a", addExpr.Left.(*ast.Variable).Name)

		mulExpr, ok := addExpr.Right.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected right side of addition to be multiplication")
		if !ok {
			return
		}
		assert.Equal(t, "*", mulExpr.Op)

		assert.IsType(t, &ast.Variable{}, mulExpr.Left)
		assert.Equal(t, "b", mulExpr.Left.(*ast.Variable).Name)
		assert.IsType(t, &ast.Variable{}, mulExpr.Right)
		assert.Equal(t, "c", mulExpr.Right.(*ast.Variable).Name)
	})

	t.Run("Parentheses", func(t *testing.T) {
		input := "(a + b) * c" // Should parse as (a + b) * c
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)
		expr, parseErr := parser.ParseExpression()

		assert.Empty(t, parser.Errors())
		assert.NoError(t, parseErr)
		assert.NotNil(t, expr)

		// Expect: BinaryExpr{Op: "*", Left: BinaryExpr{Op: "+", Left: "a", Right: "b"}, Right: "c"}
		mulExpr, ok := expr.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected outer expression to be multiplication")
		if !ok {
			return
		}
		assert.Equal(t, "*", mulExpr.Op)

		addExpr, ok := mulExpr.Left.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected left side of multiplication to be addition")
		if !ok {
			return
		}
		assert.Equal(t, "+", addExpr.Op)

		assert.IsType(t, &ast.Variable{}, addExpr.Left)
		assert.Equal(t, "a", addExpr.Left.(*ast.Variable).Name)
		assert.IsType(t, &ast.Variable{}, addExpr.Right)
		assert.Equal(t, "b", addExpr.Right.(*ast.Variable).Name)

		assert.IsType(t, &ast.Variable{}, mulExpr.Right)
		assert.Equal(t, "c", mulExpr.Right.(*ast.Variable).Name)
	})

	t.Run("Unary Minus", func(t *testing.T) {
		input := "-a * 5" // Should parse as (-a) * 5 -> (-1 * a) * 5
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)
		expr, parseErr := parser.ParseExpression()

		assert.Empty(t, parser.Errors())
		assert.NoError(t, parseErr)
		assert.NotNil(t, expr)

		// Expect: BinaryExpr{Op: "*", Left: BinaryExpr{Op: "*", Left: -1, Right: "a"}, Right: 5}
		outerMulExpr, ok := expr.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected outer expression to be multiplication")
		if !ok {
			return
		}
		assert.Equal(t, "*", outerMulExpr.Op)

		innerMulExpr, ok := outerMulExpr.Left.(*ast.BinaryExpr)
		assert.True(t, ok, "Expected left operand to be inner multiplication")
		if !ok {
			return
		}
		assert.Equal(t, "*", innerMulExpr.Op)

		numLit, ok := innerMulExpr.Left.(*ast.NumberLiteral)
		assert.True(t, ok, "Expected left operand of inner mul to be NumberLiteral")
		if ok {
			assert.Equal(t, -1.0, numLit.Value)
		}
		varLit, ok := innerMulExpr.Right.(*ast.Variable)
		assert.True(t, ok, "Expected right operand of inner mul to be Variable")
		if ok {
			assert.Equal(t, "a", varLit.Name)
		}

		numLitRight, ok := outerMulExpr.Right.(*ast.NumberLiteral)
		assert.True(t, ok, "Expected right operand of outer mul to be NumberLiteral")
		if ok {
			assert.Equal(t, 5.0, numLitRight.Value)
		}
	})

	t.Run("Function Call - sqrt", func(t *testing.T) {
		input := `\sqrt{x}`
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)
		expr, parseErr := parser.ParseExpression()

		assert.Empty(t, parser.Errors())
		assert.NoError(t, parseErr)
		assert.NotNil(t, expr)

		funcCall, ok := expr.(*ast.FuncCall)
		assert.True(t, ok, "Expected a FuncCall")
		if !ok {
			return
		}

		assert.Equal(t, "sqrt", funcCall.FuncName)
		assert.Len(t, funcCall.Args, 1, "Expected 1 argument for sqrt")
		if len(funcCall.Args) == 1 {
			argVar, okArg := funcCall.Args[0].(*ast.Variable)
			assert.True(t, okArg, "Expected argument to be a Variable")
			if okArg {
				assert.Equal(t, "x", argVar.Name)
			}
		}
	})

	t.Run("Function Call - frac", func(t *testing.T) {
		input := `\frac{a}{b}`
		lexer := NewLexer(input)
		parser := newStatefulParser(lexer)
		expr, parseErr := parser.ParseExpression()

		assert.Empty(t, parser.Errors())
		assert.NoError(t, parseErr)
		assert.NotNil(t, expr)

		funcCall, ok := expr.(*ast.FuncCall)
		assert.True(t, ok, "Expected a FuncCall")
		if !ok {
			return
		}

		assert.Equal(t, "frac", funcCall.FuncName)
		assert.Len(t, funcCall.Args, 2, "Expected 2 arguments for frac")
		if len(funcCall.Args) == 2 {
			arg1Var, okArg1 := funcCall.Args[0].(*ast.Variable)
			assert.True(t, okArg1, "Expected arg 1 to be a Variable")
			if okArg1 {
				assert.Equal(t, "a", arg1Var.Name)
			}
			arg2Var, okArg2 := funcCall.Args[1].(*ast.Variable)
			assert.True(t, okArg2, "Expected arg 2 to be a Variable")
			if okArg2 {
				assert.Equal(t, "b", arg2Var.Name)
			}
		}
	})

	// TODO: Add error cases (invalid syntax)
}
