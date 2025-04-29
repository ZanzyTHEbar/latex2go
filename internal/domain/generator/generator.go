package generator

import (
	"fmt"
	"go/format"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
)

// Generator converts internal AST Expr into Go code.
type Generator struct{}

// NewGenerator creates a fresh Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// generateExpr renders an AST expression or loop into Go code snippet.
// It also returns a boolean indicating if the generated code requires the "math" package.
func (g *Generator) generateExpr(e ast.Expr) (string, bool) {
	switch node := e.(type) {
	case *ast.NumberLiteral:
		return fmt.Sprintf("%g", node.Value), false
	case *ast.Variable:
		return node.Name, false
	case *ast.BinaryExpr:
		leftCode, leftNeedsMath := g.generateExpr(node.Left)
		rightCode, rightNeedsMath := g.generateExpr(node.Right)
		needsMath := leftNeedsMath || rightNeedsMath
		if node.Op == "^" {
			return fmt.Sprintf("math.Pow(%s, %s)", leftCode, rightCode), true // math.Pow requires math
		}
		return fmt.Sprintf("%s %s %s", leftCode, node.Op, rightCode), needsMath
	case *ast.FuncCall:
		// Special handling for frac
		if node.FuncName == "frac" {
			if len(node.Args) != 2 {
				// This should ideally be caught by the parser, but double-check here.
				return "", false // Or return an error
			}
			numeratorCode, numNeedsMath := g.generateExpr(node.Args[0])
			denominatorCode, denNeedsMath := g.generateExpr(node.Args[1])
			return fmt.Sprintf("(%s) / (%s)", numeratorCode, denominatorCode), numNeedsMath || denNeedsMath // Use parentheses for safety
		}

		// General function call handling (maps to math package)
		args := make([]string, len(node.Args))
		needsMath := false
		for i, arg := range node.Args {
			argCode, argNeedsMath := g.generateExpr(arg)
			args[i] = argCode
			needsMath = needsMath || argNeedsMath
		}

		// Check if the function is supported in the math package
		goFuncName := cases.Title(language.English, cases.Compact).String(node.FuncName)
		supportedMathFuncs := map[string]bool{"Sqrt": true, "Sin": true, "Cos": true, "Tan": true, "Pow": true /* Add others as needed */} // Pow handled by BinaryExpr ^
		if _, supported := supportedMathFuncs[goFuncName]; !supported && node.FuncName != "pow" { // Allow pow implicitly via ^
			// Return an error instead of generating invalid code
			// Note: We don't return the error directly from here, let Generate handle it.
			// For now, return empty string and signal no math needed, Generate will catch the error later.
			// TODO: A better approach might be to return an error tuple: (string, bool, error)
			return fmt.Sprintf("/* unsupported function: %s */", node.FuncName), false
		}

		// Assume math needed for all other supported func calls
		return fmt.Sprintf("math.%s(%s)",
			goFuncName,
			strings.Join(args, ", "),
		), true
	case *ast.DerivativeExpr:
		// For derivatives, we'll implement a simple finite difference approximation
		// TODO: This is a placeholder for a more sophisticated numerical differentiation, ideally using an inteface for adapters.
		bodyCode, _ := g.generateExpr(node.Body)
		
		// Implement numerical differentiation using central difference formula
		derivCode := []string{
			"func() float64 {",
			"    // Numerical differentiation using central difference",
			"    h := 0.0001 // Small step size",
		}
		
		if node.Order == 1 {
			// First-order derivative using central difference: f'(x) ≈ (f(x+h) - f(x-h)) / (2h)
			derivCode = append(derivCode,
				fmt.Sprintf("    %s := %s // Original point", node.Var, node.Var), // Assume variable is in scope
				fmt.Sprintf("    fwd := func() float64 { %s := %s + h; return %s; }() // f(x+h)", node.Var, node.Var, bodyCode),
				fmt.Sprintf("    bwd := func() float64 { %s := %s - h; return %s; }() // f(x-h)", node.Var, node.Var, bodyCode),
				"    return (fwd - bwd) / (2.0 * h)",
			)
		} else if node.Order == 2 {
			// Second-order derivative using central difference: f''(x) ≈ (f(x+h) - 2f(x) + f(x-h)) / h²
			derivCode = append(derivCode,
				fmt.Sprintf("    %s := %s // Original point", node.Var, node.Var), // Assume variable is in scope
				fmt.Sprintf("    fwd := func() float64 { %s := %s + h; return %s; }() // f(x+h)", node.Var, node.Var, bodyCode),
				fmt.Sprintf("    ctr := %s // f(x)", bodyCode),
				fmt.Sprintf("    bwd := func() float64 { %s := %s - h; return %s; }() // f(x-h)", node.Var, node.Var, bodyCode),
				"    return (fwd - 2.0*ctr + bwd) / (h * h)",
			)
		} else {
			// For higher-order derivatives, we'll just return a comment
			derivCode = append(derivCode,
				"    // Higher-order derivatives not supported",
				"    return 0.0",
			)
		}
		
		derivCode = append(derivCode, "}()")
		return strings.Join(derivCode, "\n"), true // Always needs math for numerical methods
		
	case *ast.PiecewiseExpr:
		// Generate code for piecewise function using if-else statements
		needsMath := false
		
		// Start with a function wrapper for cleaner code
		piecewiseCode := []string{
			"func() float64 {",
		}
		
		// Generate if-else statements for each case
		for i, caseItem := range node.Cases {
			valueCode, valueNeedsMath := g.generateExpr(caseItem.Value)
			needsMath = needsMath || valueNeedsMath
			
			if caseItem.Condition == nil {
				// This is the default case (otherwise/else)
				if i == len(node.Cases)-1 {
					// Last case without a condition is the default case
					piecewiseCode = append(piecewiseCode, 
						"    // Default case",
						fmt.Sprintf("    return %s", valueCode),
					)
				} else {
					// Error: cases without conditions should be last
					piecewiseCode = append(piecewiseCode, 
						"    // ERROR: Unconditional case not at end",
						fmt.Sprintf("    return %s", valueCode),
					)
				}
			} else {
				// This is a conditional case
				conditionCode, condNeedsMath := g.generateExpr(caseItem.Condition)
				needsMath = needsMath || condNeedsMath
				
				if i == 0 {
					// First condition uses "if"
					piecewiseCode = append(piecewiseCode, 
						fmt.Sprintf("    if %s {", conditionCode),
						fmt.Sprintf("        return %s", valueCode),
						"    }",
					)
				} else {
					// Subsequent conditions use "else if"
					piecewiseCode = append(piecewiseCode, 
						fmt.Sprintf("    else if %s {", conditionCode),
						fmt.Sprintf("        return %s", valueCode),
						"    }",
					)
				}
			}
		}
		
		// If no default case was provided, add one that returns NaN
		lastCase := node.Cases[len(node.Cases)-1]
		if lastCase.Condition != nil {
			piecewiseCode = append(piecewiseCode, 
				"    // No default case provided, returning NaN",
				"    return math.NaN()",
			)
			needsMath = true // Using math.NaN requires math package
		}
		
		// Close the function and call it
		piecewiseCode = append(piecewiseCode, "}()")
		
		return strings.Join(piecewiseCode, "\n"), needsMath

	case *ast.LimitExpr:
		// For limits, we'll implement a simple approximation by evaluating at a point very close to the limit
		bodyCode, bodyNeedsMath := g.generateExpr(node.Body)
		approachesCode, approachesNeedsMath := g.generateExpr(node.Approaches)
		
		// Implementation approach: evaluate at a point very close to the limit
		limitCode := []string{
			"func() float64 {",
			"    // Approximating limit by evaluating at a point very close to the target",
			"    epsilon := 1e-10 // Small value for approximation",
			fmt.Sprintf("    target := %s // Value approached", approachesCode),
			fmt.Sprintf("    %s := float64(target) + epsilon // Set variable slightly above target", node.Var),
			fmt.Sprintf("    return %s // Evaluate expression", bodyCode),
			"}()",
		}
		
		return strings.Join(limitCode, "\n"), bodyNeedsMath || approachesNeedsMath

	case *ast.IntegralExpr:
		// For integrals, we'll use numerical integration based on the trapezoidal rule
		// For definite integrals, we can implement basic numerical integration
		bodyCode, bodyNeedsMath := g.generateExpr(node.Body)
		
		if node.IsDefinite {
			// Generate definite integral using numerical integration
			lowerCode, lowerNeedsMath := g.generateExpr(node.Lower)
			upperCode, upperNeedsMath := g.generateExpr(node.Upper)
			
			// We need to implement a basic numerical integration algorithm
			// Using the trapezoidal rule for simplicity
			integralCode := []string{
				"func() float64 {",
				fmt.Sprintf("    a := %s // Lower bound", lowerCode),
				fmt.Sprintf("    b := %s // Upper bound", upperCode),
				"    n := 1000 // Number of intervals for numerical integration",
				"    h := (b - a) / float64(n)",
				"    sum := 0.0",
				"    for i := 0; i <= n; i++ {",
				fmt.Sprintf("        %s := a + float64(i)*h // Integration variable", node.Var),
				fmt.Sprintf("        fx := %s // Integrand", bodyCode),
				"        weight := 1.0",
				"        if i == 0 || i == n {",
				"            weight = 0.5",
				"        }",
				"        sum += weight * fx",
				"    }",
				"    return sum * h",
				"}()",
			}
			
			return strings.Join(integralCode, "\n"), bodyNeedsMath || lowerNeedsMath || upperNeedsMath
		} else {
			// For indefinite integrals, we can only return a comment as symbolic integration
			// is beyond the scope of a simple translator
			// TODO: Implement a more sophisticated symbolic integration approach
			return fmt.Sprintf("/* Symbolic integration of %s with respect to %s not supported */", 
				bodyCode, node.Var), bodyNeedsMath
		}

	case *ast.FactorialExpr:
		// Generate factorial using math.Gamma(n+1)
		valueCode, _ := g.generateExpr(node.Value)
		// Use math.Gamma(x+1) for factorial calculation
		return fmt.Sprintf("math.Gamma(%s + 1.0)", valueCode), true

	case *ast.SumExpr:
		// Summation or product loop
		idx := node.Var
		lowCode, lowNeedsMath := g.generateExpr(node.Lower)
		upCode, upNeedsMath := g.generateExpr(node.Upper)
		bodyCode, bodyNeedsMath := g.generateExpr(node.Body)
		needsMath := lowNeedsMath || upNeedsMath || bodyNeedsMath

		initVal, op := "0.0", "+" // Use float literal for init
		if node.IsProduct {
			initVal, op = "1.0", "*"
		}
		// Ensure loop bounds are treated as floats for comparison if they are variables
		// Note: This assumes loop variables are integers, which might be fragile.
		// TODO: A more robust solution might involve type analysis or clearer loop semantics.
		loop := []string{
			fmt.Sprintf("result := %s", initVal),
			// Using float64 for loop counter and bounds for consistency with math ops
			fmt.Sprintf("for %s := float64(int(%s)); %s <= float64(int(%s)); %s++ {", idx, lowCode, idx, upCode, idx),
			fmt.Sprintf("    result = result %s (%s)", op, bodyCode), // Add parentheses around body for safety
			"}",
			"return result", // Return result directly from loop structure
		}
		return strings.Join(loop, "\n"), needsMath
	default:
		return "", false
	}
}

// Generate produces full Go source code for the given AST root, package, and function.
func (g *Generator) Generate(root ast.Expr, pkgName, funcName string) (string, error) {
	// Generate the core expression/loop code and check if math is needed
	codeBody, needsMath := g.generateExpr(root)

	// Check for unsupported function placeholder generated by generateExpr
	if strings.HasPrefix(codeBody, "/* unsupported function:") {
		var unsupportedFuncName string
		fmt.Sscanf(codeBody, "/* unsupported function: %s */", &unsupportedFuncName)
		return "", fmt.Errorf("unsupported LaTeX function: %s", unsupportedFuncName)
	}

	mathImport := ""
	if needsMath {
		mathImport = "\"math\""
	}

	var header string
	if mathImport != "" {
		header = fmt.Sprintf("package %s\n\nimport %s\n\n", pkgName, mathImport)
	} else {
		header = fmt.Sprintf("package %s\n\n", pkgName)
	}

	// Collect variables from AST
	vars := make(map[string]struct{})
	var collect func(e ast.Expr, loopVar string) // Pass loopVar down
	collect = func(e ast.Expr, loopVar string) {
		if e == nil { // Add nil check for safety
			return
		}
		switch n := e.(type) {
		case *ast.Variable:
			// Exclude loop variable from parameters
			if n.Name != loopVar {
				vars[sanitizeVariableName(n.Name)] = struct{}{}
			}
		case *ast.BinaryExpr:
			collect(n.Left, loopVar)
			collect(n.Right, loopVar)
		case *ast.FuncCall:
			// Don't collect from inside frac if it was handled specially
			if n.FuncName != "frac" {
				for _, a := range n.Args {
					collect(a, loopVar)
				}
			} else {
				// Need to collect from frac args manually if handled specially
				if len(n.Args) == 2 {
					collect(n.Args[0], loopVar)
					collect(n.Args[1], loopVar)
				}
			}
		case *ast.SumExpr:
			// Collect from bounds, passing the current loopVar (if any)
			collect(n.Lower, loopVar)
			collect(n.Upper, loopVar)
			// Collect from body, passing the *new* loopVar for this SumExpr
			collect(n.Body, n.Var)
		case *ast.IntegralExpr:
			// Collect from bounds for definite integrals
			if n.IsDefinite {
				collect(n.Lower, loopVar)
				collect(n.Upper, loopVar)
			}
			// Collect from body, passing the integration variable as loopVar to exclude it
			collect(n.Body, n.Var)
		case *ast.DerivativeExpr:
			// Collect from body, passing the differentiation variable as loopVar
			collect(n.Body, n.Var)
		case *ast.LimitExpr:
			// Collect from approaches value
			collect(n.Approaches, loopVar)
			// Collect from body, passing the limit variable as loopVar
			collect(n.Body, n.Var)
		case *ast.FactorialExpr:
			// Collect from the factorial's value
			collect(n.Value, loopVar)
		case *ast.PiecewiseExpr:
			// Collect from all case values and conditions
			for _, caseItem := range n.Cases {
				collect(caseItem.Value, loopVar)
				if caseItem.Condition != nil {
					collect(caseItem.Condition, loopVar)
				}
			}
		}
	}
	collect(root, "") // Start collection with no loop variable context

	// Build sorted parameter list
	names := make([]string, 0, len(vars))
	for v := range vars {
		names = append(names, v)
	}
	sort.Strings(names)
	params := ""
	if len(names) > 0 {
		parts := make([]string, len(names))
		for i, v := range names { // Corrected loop syntax
			parts[i] = fmt.Sprintf("%s float64", v) // Use sanitized name
		}
		params = strings.Join(parts, ", ")
	}

	// Assemble the function body
	var funcBody string
	if _, ok := root.(*ast.SumExpr); ok {
		// For SumExpr, the generateExpr already returns the full loop and return statement
		indented := indent(codeBody, "\t")
		funcBody = fmt.Sprintf("func %s(%s) float64 {\n%s\n}", funcName, params, indented)
	} else {
		// For simple expressions, add the return statement
		funcBody = fmt.Sprintf("func %s(%s) float64 {\n\treturn %s\n}", funcName, params, codeBody)
	}

	src := header + funcBody

	// Format with go/format
	formatted, err := format.Source([]byte(src))
	if err != nil {
		// If formatting fails, return the unformatted source and the error for debugging
		return src, fmt.Errorf("failed to format generated code: %w\nSource:\n%s", err, src)
	}
	return string(formatted), nil
}

// indent prefixes each line of s with prefix.
func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

// goKeywords is a set of Go reserved keywords.
var goKeywords = map[string]struct{}{
	"break": {}, "default": {}, "func": {}, "interface": {}, "select": {},
	"case": {}, "defer": {}, "go": {}, "map": {}, "struct": {},
	"chan": {}, "else": {}, "goto": {}, "package": {}, "switch": {},
	"const": {}, "fallthrough": {}, "if": {}, "range": {}, "type": {},
	"continue": {}, "for": {}, "import": {}, "return": {}, "var": {},
	// Technically not keywords, but often problematic as variable names
	"true": {}, "false": {}, "nil": {}, "iota": {},
}

// sanitizeVariableName checks if a name is a Go keyword and appends an underscore if it is.
func sanitizeVariableName(name string) string {
	if _, isKeyword := goKeywords[name]; isKeyword {
		return name + "_"
	}
	return name
}
