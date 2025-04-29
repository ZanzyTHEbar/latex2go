package ast

// Node represents any node in the equation's abstract syntax tree.
// It serves as a marker interface for all AST node types.
type Node interface {
	node() // Internal marker method
}

// Expr represents an expression node within the AST.
// Expressions evaluate to a value (e.g., numbers, variables, operations).
type Expr interface {
	Node
	expr() // Internal marker method
}

// --- Concrete Node Types ---

// NumberLiteral represents a numeric value (e.g., 3.14, 42).
type NumberLiteral struct {
	Value float64
}

func (NumberLiteral) node() {}
func (NumberLiteral) expr() {}

// Variable represents a variable identifier (e.g., x, y, a).
type Variable struct {
	Name string
}

func (Variable) node() {}
func (Variable) expr() {}

// BinaryExpr represents an operation with two operands (e.g., a + b, x ^ 2).
type BinaryExpr struct {
	Op    string // Operator token (e.g., "+", "-", "*", "/", "^")
	Left  Expr   // Left-hand side expression
	Right Expr   // Right-hand side expression
}

func (BinaryExpr) node() {}
func (BinaryExpr) expr() {}

// FuncCall represents a function call (e.g., \sqrt{x}, \sin{y}, \frac{a}{b}).
// Note: \frac{a}{b} is treated like a function call in this AST,
// the generator will handle its specific translation to Go division.
type FuncCall struct {
	FuncName string // LaTeX command name (e.g., "sqrt", "sin", "cos", "frac")
	Args     []Expr // Arguments provided to the function/command
}

func (FuncCall) node() {}
func (FuncCall) expr() {}

// SumExpr represents a summation or product (e.g., \sum_{i=1}^{n} f(i), \prod_{i=1}^{n} f(i)).
type SumExpr struct {
	IsProduct   bool   // true for product (\prod), false for sum (\sum)
	Var         string // Summation variable (e.g., "i")
	Lower, Upper Expr  // Lower and upper bounds (e.g., 1, n)
	Body        Expr   // The expression to sum/product over (e.g., f(i))
}

func (SumExpr) node() {}
func (SumExpr) expr() {}

// TODO: Add IntegralExpr, DerivativeExpr, LimitExpr, PiecewiseExpr, SetIterationExpr as needed.
