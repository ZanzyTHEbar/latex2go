package app

import (
	// Import domain components used in interfaces
	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
)

// Config holds configuration values passed from the input adapter.
type Config struct {
	OutputFile  string
	PackageName string
	FuncName    string
}

// LatexProvider defines the input port for retrieving LaTeX input and config.
type LatexProvider interface {
	GetLatexInput() (latex string, config Config, err error)
}

// GoCodeWriter defines the output port for writing the generated Go code.
type GoCodeWriter interface {
	WriteGoCode(code string) error
}

// --- Domain Service Interfaces ---
// These interfaces define the contracts for domain services used by the application.

// Parser defines the input port for parsing LaTeX input into an AST.
type Parser interface {
	Parse(latexString string) (ast.Expr, error)
}

// Generator defines the output port for generating Go code from an AST.
type Generator interface {
	Generate(root ast.Expr, pkgName, funcName string) (string, error)
}
