package generator

import (
	"bytes"
	"fmt"
	goast "go/ast" // Alias standard Go AST package
	"go/format"
	"go/token"
	"sort"
	"strings"

	// "github.com/ZanzyTHEbar/latex2go/internal/app" // REMOVED to break import cycle
	internalast "github.com/ZanzyTHEbar/latex2go/internal/domain/ast" // Alias our internal AST
)

// Generator converts our internal AST into Go code.
type Generator struct {
	// Potentially add configuration or state if needed later
}

// NewGenerator creates a new code generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate takes the root of our internal expression AST and configuration parameters,
// and returns a string containing the formatted Go code.
func (g *Generator) Generate(root internalast.Expr, packageName, funcName string) (string, error) {
	// 1. Collect variables from the AST to determine function parameters.
	vars := make(map[string]struct{})
	collectVariables(root, vars)
	sortedVars := make([]string, 0, len(vars))
	for v := range vars {
		sortedVars = append(sortedVars, v)
	}
	sort.Strings(sortedVars) // Ensure consistent parameter order

	// 2. Build the Go AST expression tree from our internal AST.
	goExpr, err := g.buildGoExpr(root)
	if err != nil {
		return "", fmt.Errorf("failed to build Go expression AST: %w", err)
	}

	// 3. Create function parameters.
	params := &goast.FieldList{List: make([]*goast.Field, len(sortedVars))}
	for i, varName := range sortedVars {
		params.List[i] = &goast.Field{
			Names: []*goast.Ident{goast.NewIdent(varName)},
			Type:  goast.NewIdent("float64"), // Assume float64 for all variables
		}
	}

	// 4. Create function return type.
	results := &goast.FieldList{
		List: []*goast.Field{
			{
				Type: goast.NewIdent("float64"), // Assume float64 return type
			},
		},
	}

	// 5. Create the function body with a return statement.
	funcBody := &goast.BlockStmt{
		List: []goast.Stmt{
			&goast.ReturnStmt{Results: []goast.Expr{goExpr}},
		},
	}

	// 6. Create the function declaration.
	funcDecl := &goast.FuncDecl{
		Name: goast.NewIdent(funcName), // Use funcName argument
		Type: &goast.FuncType{
			Params:  params,
			Results: results,
		},
		Body: funcBody,
	}

	// 7. Create the file AST.
	file := &goast.File{
		Name: goast.NewIdent(packageName), // Use packageName argument
		Decls: []goast.Decl{
			// Add import declaration for "math" if needed (check during buildGoExpr)
			&goast.GenDecl{
				Tok: token.IMPORT,
				Specs: []goast.Spec{
					&goast.ImportSpec{Path: &goast.BasicLit{Kind: token.STRING, Value: `"math"`}},
				},
			},
			funcDecl,
		},
	}

	// 8. Format the Go AST into source code.
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), file); err != nil {
		return "", fmt.Errorf("failed to format generated Go code: %w", err)
	}

	return buf.String(), nil
}

// collectVariables recursively traverses the AST and collects unique variable names.
func collectVariables(node internalast.Node, vars map[string]struct{}) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *internalast.Variable:
		vars[n.Name] = struct{}{}
	case *internalast.BinaryExpr:
		collectVariables(n.Left, vars)
		collectVariables(n.Right, vars)
	case *internalast.FuncCall:
		for _, arg := range n.Args {
			collectVariables(arg, vars)
		}
	case *internalast.NumberLiteral:
	// Numbers are not variables
	default:
		// Should not happen with current AST structure
		fmt.Printf("Warning: Unhandled node type in collectVariables: %T\n", n)
	}
}

// buildGoExpr recursively translates our internal AST expression nodes
// into corresponding goast expression nodes.
func (g *Generator) buildGoExpr(node internalast.Expr) (goast.Expr, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot build Go expression from nil internal AST node")
	}

	switch n := node.(type) {
	case *internalast.NumberLiteral:
		// Represent as float literal in Go
		return &goast.BasicLit{Kind: token.FLOAT, Value: fmt.Sprintf("%f", n.Value)}, nil

	case *internalast.Variable:
		// Represent as an identifier
		return goast.NewIdent(n.Name), nil

	case *internalast.BinaryExpr:
		leftExpr, err := g.buildGoExpr(n.Left)
		if err != nil {
			return nil, err
		}
		rightExpr, err := g.buildGoExpr(n.Right)
		if err != nil {
			return nil, err
		}

		var goOp token.Token
		switch n.Op {
		case "+":
			goOp = token.ADD
		case "-":
			goOp = token.SUB
		case "*":
			goOp = token.MUL
		case "/":
			goOp = token.QUO // Division
		case "^":
			// Exponentiation (a^b) translates to math.Pow(a, b)
			return &goast.CallExpr{
				Fun:  goast.NewIdent("math.Pow"), // Assumes "math" is imported
				Args: []goast.Expr{leftExpr, rightExpr},
			}, nil
		default:
			return nil, fmt.Errorf("unsupported binary operator: %s", n.Op)
		}
		return &goast.BinaryExpr{X: leftExpr, Op: goOp, Y: rightExpr}, nil

	case *internalast.FuncCall:
		args := make([]goast.Expr, len(n.Args))
		for i, arg := range n.Args {
			goArg, err := g.buildGoExpr(arg)
			if err != nil {
				return nil, fmt.Errorf("failed to build argument %d for function %s: %w", i, n.FuncName, err)
			}
			args[i] = goArg
		}

		// Map LaTeX function names to Go math function names
		var goFuncName string
		switch strings.ToLower(n.FuncName) {
		case "sqrt":
			goFuncName = "math.Sqrt"
		case "sin":
			goFuncName = "math.Sin"
		case "cos":
			goFuncName = "math.Cos"
		case "tan":
			goFuncName = "math.Tan"
		// Special case: \frac{a}{b} becomes division
		case "frac":
			if len(args) != 2 {
				return nil, fmt.Errorf("frac function requires exactly 2 arguments, got %d", len(args))
			}
			// Translate \frac{a}{b} to a / b
			return &goast.BinaryExpr{X: args[0], Op: token.QUO, Y: args[1]}, nil
		default:
			return nil, fmt.Errorf("unsupported LaTeX function: %s", n.FuncName)
		}

		return &goast.CallExpr{
			Fun:  goast.NewIdent(goFuncName), // Assumes "math" is imported
			Args: args,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported internal AST node type: %T", n)
	}
}
