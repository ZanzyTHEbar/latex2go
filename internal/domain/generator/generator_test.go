package generator

import (
	"testing"

	"go/parser"
	"go/token"

	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast" // Use correct import path
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	gen := NewGenerator()

	t.Run("Simple Addition", func(t *testing.T) {
		// Input AST for "a + b"
		inputAST := &ast.BinaryExpr{
			Op:    "+",
			Left:  &ast.Variable{Name: "a"},
			Right: &ast.Variable{Name: "b"},
		}

		// Generate Go code
		// Provide packageName ("main") as the second argument
		goCode, err := gen.Generate(inputAST, "main", "generatedFunc")

		// Assertions
		assert.NoError(t, err, "Generator returned an error")
		require.NotEmpty(t, goCode, "Generator returned empty code")

		// Optional: Parse the generated code to check its validity
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code is not valid Go code:\n%s", goCode)

		// TODO: Add more specific assertions about the generated code structure
		// This might involve parsing the expected and actual code and comparing ASTs,
		// or using string comparisons after normalizing whitespace/formatting.
		// For now, we check for basic validity and presence of key elements.
		// Check signature parts separately to be less sensitive to spacing.
		assert.Contains(t, goCode, "func generatedFunc(", "Function definition start missing")
		assert.Contains(t, goCode, "a float64", "Parameter 'a' missing or incorrect type")
		assert.Contains(t, goCode, "b float64", "Parameter 'b' missing or incorrect type")
		assert.Contains(t, goCode, ") float64", "Return type missing or incorrect")
		assert.Contains(t, goCode, "return a + b", "Return statement mismatch")
		assert.Contains(t, goCode, `import "math"`, "Math import missing") // Generator always adds it currently

	})

	t.Run("Subtraction", func(t *testing.T) {
		inputAST := &ast.BinaryExpr{Op: "-", Left: &ast.Variable{Name: "x"}, Right: &ast.NumberLiteral{Value: 3.14}}
		goCode, err := gen.Generate(inputAST, "main", "subFunc")
		assert.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code invalid:\n%s", goCode)
		assert.Contains(t, goCode, "func subFunc(x float64) float64") // Simpler signature check
		assert.Contains(t, goCode, "return x - 3.14")                 // Allow for float formatting variations
	})

	t.Run("Multiplication and Division", func(t *testing.T) {
		// AST for (a * b) / c
		inputAST := &ast.BinaryExpr{
			Op: "/",
			Left: &ast.BinaryExpr{
				Op:    "*",
				Left:  &ast.Variable{Name: "a"},
				Right: &ast.Variable{Name: "b"},
			},
			Right: &ast.Variable{Name: "c"},
		}
		goCode, err := gen.Generate(inputAST, "main", "calcFunc")
		assert.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code invalid:\n%s", goCode)
		// Check signature parts
		assert.Contains(t, goCode, "func calcFunc(")
		assert.Contains(t, goCode, "a float64")
		assert.Contains(t, goCode, "b float64")
		assert.Contains(t, goCode, "c float64")
		assert.Contains(t, goCode, ") float64")
		// Check for the core expression - Go formatting might add parentheses
		assert.Contains(t, goCode, "a * b / c")
	})

	t.Run("Exponentiation", func(t *testing.T) {
		// AST for a ^ 2
		inputAST := &ast.BinaryExpr{
			Op:    "^",
			Left:  &ast.Variable{Name: "a"},
			Right: &ast.NumberLiteral{Value: 2},
		}
		goCode, err := gen.Generate(inputAST, "main", "powFunc")
		assert.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code invalid:\n%s", goCode)
		assert.Contains(t, goCode, "func powFunc(a float64) float64")
		assert.Contains(t, goCode, "return math.Pow(a, 2") // Check start of Pow call, ignore exact float format
		assert.Contains(t, goCode, `import "math"`)
	})

	t.Run("Function Call - sqrt", func(t *testing.T) {
		// AST for \sqrt{x}
		inputAST := &ast.FuncCall{
			FuncName: "sqrt",
			Args:     []ast.Expr{&ast.Variable{Name: "x"}},
		}
		goCode, err := gen.Generate(inputAST, "main", "sqrtFunc")
		assert.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code invalid:\n%s", goCode)
		assert.Contains(t, goCode, "func sqrtFunc(x float64) float64")
		assert.Contains(t, goCode, "return math.Sqrt(x)")
		assert.Contains(t, goCode, `import "math"`)
	})

	t.Run("Function Call - frac", func(t *testing.T) {
		// AST for \frac{a}{b}
		inputAST := &ast.FuncCall{
			FuncName: "frac",
			Args: []ast.Expr{
				&ast.Variable{Name: "a"},
				&ast.Variable{Name: "b"},
			},
		}
		goCode, err := gen.Generate(inputAST, "main", "fracFunc")
		assert.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code invalid:\n%s", goCode)
		// Check signature parts
		assert.Contains(t, goCode, "func fracFunc(")
		assert.Contains(t, goCode, "a float64")
		assert.Contains(t, goCode, "b float64")
		assert.Contains(t, goCode, ") float64")
		assert.Contains(t, goCode, "return a / b") // frac translates to division
		assert.Contains(t, goCode, `import "math"`)
	})

	t.Run("Function Call - sin", func(t *testing.T) {
		// AST for \sin{x}
		inputAST := &ast.FuncCall{
			FuncName: "sin",
			Args:     []ast.Expr{&ast.Variable{Name: "x"}},
		}
		goCode, err := gen.Generate(inputAST, "main", "sinFunc")
		assert.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code invalid:\n%s", goCode)
		assert.Contains(t, goCode, "func sinFunc(x float64) float64")
		assert.Contains(t, goCode, "return math.Sin(x)")
		assert.Contains(t, goCode, `import "math"`)
	})

	t.Run("Complex Expression", func(t *testing.T) {
		// AST for \frac{-b + \sqrt{b^2 - 4*a*c}}{2*a}
		inputAST := &ast.FuncCall{
			FuncName: "frac",
			Args: []ast.Expr{
				&ast.BinaryExpr{ // -b + sqrt(...)
					Op: "+",
					Left: &ast.BinaryExpr{ // -b -> -1 * b
						Op:    "*",
						Left:  &ast.NumberLiteral{Value: -1},
						Right: &ast.Variable{Name: "b"},
					},
					Right: &ast.FuncCall{ // sqrt(...)
						FuncName: "sqrt",
						Args: []ast.Expr{
							&ast.BinaryExpr{ // b^2 - 4*a*c
								Op: "-",
								Left: &ast.BinaryExpr{ // b^2
									Op:    "^",
									Left:  &ast.Variable{Name: "b"},
									Right: &ast.NumberLiteral{Value: 2},
								},
								Right: &ast.BinaryExpr{ // 4*a*c -> (4*a)*c
									Op: "*",
									Left: &ast.BinaryExpr{ // 4*a
										Op:    "*",
										Left:  &ast.NumberLiteral{Value: 4},
										Right: &ast.Variable{Name: "a"},
									},
									Right: &ast.Variable{Name: "c"},
								},
							},
						},
					},
				},
				&ast.BinaryExpr{ // 2*a
					Op:    "*",
					Left:  &ast.NumberLiteral{Value: 2},
					Right: &ast.Variable{Name: "a"},
				},
			},
		}

		goCode, err := gen.Generate(inputAST, "main", "quadraticFormulaPart")
		assert.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
		assert.NoError(t, err, "Generated code invalid:\n%s", goCode)

		// Check signature parts
		assert.Contains(t, goCode, "func quadraticFormulaPart(")
		assert.Contains(t, goCode, "a float64")
		assert.Contains(t, goCode, "b float64")
		assert.Contains(t, goCode, "c float64")
		assert.Contains(t, goCode, ") float64")
		// Check for key parts, acknowledging formatting might vary
		assert.Contains(t, goCode, "math.Sqrt")
		assert.Contains(t, goCode, "math.Pow")
		assert.Contains(t, goCode, "/") // From frac and potentially internal division
		assert.Contains(t, goCode, "*")
		assert.Contains(t, goCode, "+")
		assert.Contains(t, goCode, "-")
		assert.Contains(t, goCode, `import "math"`)
	})

}
