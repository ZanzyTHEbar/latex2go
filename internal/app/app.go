package app

import (
	"fmt"

	"github.com/ZanzyTHEbar/latex2go/internal/domain/generator"
	"github.com/ZanzyTHEbar/latex2go/internal/domain/parser"
)

// Latex2GoService defines the core application service.
type Latex2GoService struct {
	parser    *parser.Parser
	generator *generator.Generator
}

// NewLatex2GoService creates a new instance of the application service.
func NewLatex2GoService(p *parser.Parser, g *generator.Generator) *Latex2GoService {
	return &Latex2GoService{
		parser:    p,
		generator: g,
	}
}

// ConvertLatexToGo takes a LaTeX string and converts it into a Go function string.
// It uses the configured parser and generator.
// funcName specifies the name for the generated Go function.
// packageName specifies the package name for the generated Go code.
func (s *Latex2GoService) ConvertLatexToGo(latexInput, packageName, funcName string) (string, error) {
	if latexInput == "" {
		return "", fmt.Errorf("latex input cannot be empty")
	}
	if packageName == "" {
		packageName = "main" // Default package name
	}
	if funcName == "" {
		funcName = "generatedFunc" // Default function name
	}

	// Parse the LaTeX input into an AST
	ast, err := s.parser.Parse(latexInput)
	if err != nil {
		return "", fmt.Errorf("parsing error: %w", err)
	}

	// Generate Go code from the AST
	goCode, err := s.generator.Generate(ast, packageName, funcName)
	if err != nil {
		return "", fmt.Errorf("code generation error: %w", err)
	}

	return goCode, nil
}
