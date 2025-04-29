package generator

import (
	"testing"

	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
	"github.com/stretchr/testify/assert"
)

func TestGenerator_AdvancedExpressions(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name          string
		expr          ast.Expr
		expectMath    bool
		expectPattern string
	}{
		{
			name:          "Factorial",
			expr:          &ast.FactorialExpr{Value: &ast.NumberLiteral{Value: 5.0}},
			expectMath:    true,
			expectPattern: "math.Gamma(5 + 1.0)",
		},
		{
			name: "Definite Integral",
			expr: &ast.IntegralExpr{
				IsDefinite: true,
				Var:        "x",
				Lower:      &ast.NumberLiteral{Value: 0.0},
				Upper:      &ast.NumberLiteral{Value: 1.0},
				Body:       &ast.Variable{Name: "x"},
			},
			expectMath:    false,
			expectPattern: "// Lower bound",
		},
		{
			name: "Derivative",
			expr: &ast.DerivativeExpr{
				IsPartial: false,
				Var:       "x",
				Order:     1,
				Body:      &ast.Variable{Name: "x"},
			},
			expectMath:    true,
			expectPattern: "central difference",
		},
		{
			name: "Limit",
			expr: &ast.LimitExpr{
				Var:        "x",
				Approaches: &ast.NumberLiteral{Value: 0.0},
				Body:       &ast.Variable{Name: "x"},
			},
			expectMath:    false,
			expectPattern: "epsilon",
		},
		{
			name: "Piecewise",
			expr: &ast.PiecewiseExpr{
				Cases: []ast.PiecewiseCase{
					{Value: &ast.NumberLiteral{Value: 1.0}, Condition: &ast.Variable{Name: "condition1"}},
					{Value: &ast.NumberLiteral{Value: 2.0}, Condition: &ast.Variable{Name: "condition2"}},
					{Value: &ast.NumberLiteral{Value: 3.0}, Condition: nil}, // Default case
				},
			},
			expectMath:    false,
			expectPattern: "if condition1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, needsMath := gen.generateExpr(tt.expr)
			assert.Equal(t, tt.expectMath, needsMath)
			assert.Contains(t, code, tt.expectPattern)
		})
	}
}
