package generator

import (
	"fmt" // Added import for fmt.Sprintf
	"strings"
	"testing"

	"go/parser"
	"go/token"

	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast" // Use correct import path
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to check generated code validity and basic structure
func checkGeneratedCode(t *testing.T, goCode string, err error, expectedPkg, expectedFunc string, expectedParams []string, requiresMath bool) {
	t.Helper()
	require.NoError(t, err, "Generator returned an error")
	require.NotEmpty(t, goCode, "Generator returned empty code")

	// Check Go syntax validity
	_, parseErr := parser.ParseFile(token.NewFileSet(), "", goCode, parser.AllErrors)
	require.NoError(t, parseErr, "Generated code is not valid Go code:\n%s", goCode)

	// Check package name
	assert.Contains(t, goCode, fmt.Sprintf("package %s", expectedPkg), "Incorrect package name")

	// Check function signature start
	sigStart := fmt.Sprintf("func %s(", expectedFunc)
	assert.Contains(t, goCode, sigStart, "Function signature start mismatch")

	// Check function parameters (order matters)
	var paramStrings []string
	for _, p := range expectedParams {
		paramStrings = append(paramStrings, fmt.Sprintf("%s float64", p))
	}
	expectedSig := fmt.Sprintf("%s%s) float64", sigStart, strings.Join(paramStrings, ", "))
	assert.Contains(t, goCode, expectedSig, "Function signature mismatch (parameters/return type)")

	// Check math import
	if requiresMath {
		assert.Contains(t, goCode, `import "math"`, "Expected 'import \"math\"'")
	} else {
		 // Now that the generator logic is updated, assert that math is NOT imported.
		assert.NotContains(t, goCode, `import "math"`, "Did not expect 'import \"math\"'")
	}
}

func TestGenerator(t *testing.T) {
	gen := NewGenerator()

	t.Run("Simple Addition - No Math Import Needed", func(t *testing.T) {
		inputAST := &ast.BinaryExpr{
			Op:    "+",
			Left:  &ast.Variable{Name: "a"},
			Right: &ast.Variable{Name: "b"},
		}
		goCode, err := gen.Generate(inputAST, "main", "addFunc")
		checkGeneratedCode(t, goCode, err, "main", "addFunc", []string{"a", "b"}, false) // Expect no math needed
		assert.Contains(t, goCode, "return a + b", "Return statement mismatch")
	})

	t.Run("Variable Sorting", func(t *testing.T) {
		// Input AST for "z + y + x"
		inputAST := &ast.BinaryExpr{
			Op: "+",
			Left: &ast.BinaryExpr{
				Op:    "+",
				Left:  &ast.Variable{Name: "z"},
				Right: &ast.Variable{Name: "y"},
			},
			Right: &ast.Variable{Name: "x"},
		}
		goCode, err := gen.Generate(inputAST, "main", "sortedParamsFunc")
		// Expect parameters sorted alphabetically: x, y, z
		checkGeneratedCode(t, goCode, err, "main", "sortedParamsFunc", []string{"x", "y", "z"}, false) // Expect no math needed
		assert.Contains(t, goCode, "return z + y + x", "Return statement mismatch") // Go formatting might change order slightly, but check core elements
	})

	t.Run("Custom Package and Func Name", func(t *testing.T) {
		inputAST := &ast.BinaryExpr{Op: "*", Left: &ast.Variable{Name: "val"}, Right: &ast.NumberLiteral{Value: 2}}
		goCode, err := gen.Generate(inputAST, "custompkg", "multiplyByTwo")
		checkGeneratedCode(t, goCode, err, "custompkg", "multiplyByTwo", []string{"val"}, false) // Expect no math needed
		assert.Contains(t, goCode, "return val * 2", "Return statement mismatch")
	})

	t.Run("Subtraction", func(t *testing.T) {
		inputAST := &ast.BinaryExpr{Op: "-", Left: &ast.Variable{Name: "x"}, Right: &ast.NumberLiteral{Value: 3.14}}
		goCode, err := gen.Generate(inputAST, "main", "subFunc")
		checkGeneratedCode(t, goCode, err, "main", "subFunc", []string{"x"}, false) // Expect no math needed
		assert.Contains(t, goCode, "return x - 3.14") // Allow for float formatting variations
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
		checkGeneratedCode(t, goCode, err, "main", "calcFunc", []string{"a", "b", "c"}, false) // Expect no math needed
		// Check for the core expression - Go formatting might add parentheses
		assert.Contains(t, goCode, "a * b / c")
	})

	t.Run("Exponentiation - Requires Math", func(t *testing.T) {
		// AST for a ^ 2
		inputAST := &ast.BinaryExpr{
			Op:    "^",
			Left:  &ast.Variable{Name: "a"},
			Right: &ast.NumberLiteral{Value: 2},
		}
		goCode, err := gen.Generate(inputAST, "main", "powFunc")
		checkGeneratedCode(t, goCode, err, "main", "powFunc", []string{"a"}, true) // Expect math needed
		assert.Contains(t, goCode, "return math.Pow(a, 2") // Check start of Pow call, ignore exact float format
	})

	t.Run("Function Call - sqrt - Requires Math", func(t *testing.T) {
		// AST for \sqrt{x}
		inputAST := &ast.FuncCall{
			FuncName: "sqrt",
			Args:     []ast.Expr{&ast.Variable{Name: "x"}},
		}
		goCode, err := gen.Generate(inputAST, "main", "sqrtFunc")
		checkGeneratedCode(t, goCode, err, "main", "sqrtFunc", []string{"x"}, true) // Expect math needed
		assert.Contains(t, goCode, "return math.Sqrt(x)")
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
		checkGeneratedCode(t, goCode, err, "main", "fracFunc", []string{"a", "b"}, false) // Expect no math needed
		assert.Contains(t, goCode, "return (a) / (b)") // frac translates to division with parentheses
	})

	t.Run("Function Call - sin - Requires Math", func(t *testing.T) {
		// AST for \sin{x}
		inputAST := &ast.FuncCall{
			FuncName: "sin",
			Args:     []ast.Expr{&ast.Variable{Name: "x"}},
		}
		goCode, err := gen.Generate(inputAST, "main", "sinFunc")
		checkGeneratedCode(t, goCode, err, "main", "sinFunc", []string{"x"}, true) // Expect math needed
		assert.Contains(t, goCode, "return math.Sin(x)")
	})

	t.Run("Complex Expression - Requires Math", func(t *testing.T) {
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

		goCode, err := gen.Generate(inputAST, "mathops", "quadraticFormulaPart")
		checkGeneratedCode(t, goCode, err, "mathops", "quadraticFormulaPart", []string{"a", "b", "c"}, true) // Expect math needed

		// Check for key parts, acknowledging formatting might vary
		assert.Contains(t, goCode, "math.Sqrt")
		assert.Contains(t, goCode, "math.Pow")
		assert.Contains(t, goCode, "/") // From frac and potentially internal division
		assert.Contains(t, goCode, "*")
		assert.Contains(t, goCode, "+")
		assert.Contains(t, goCode, "-")
	})

	t.Run("Unsupported Function Error", func(t *testing.T) {
		// AST for \unknown{x}
		inputAST := &ast.FuncCall{
			FuncName: "unknown",
			Args:     []ast.Expr{&ast.Variable{Name: "x"}},
		}
		_, err := gen.Generate(inputAST, "main", "failFunc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported LaTeX function: unknown")
	})

	// TODO: Add test for unsupported AST node type if a relevant scenario exists

}
