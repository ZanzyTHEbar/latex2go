package app

import (
	"fmt"

	// Import domain components (adjust paths/names if they differ)
	"github.com/ZanzyTHEbar/latex2go/internal/domain/generator"
	"github.com/ZanzyTHEbar/latex2go/internal/domain/parser"
)

// ApplicationService orchestrates the LaTeX to Go conversion process.
type ApplicationService struct {
	latexProvider LatexProvider        // Input port
	codeWriter    GoCodeWriter         // Output port
	parser        *parser.Parser       // Domain: LaTeX parser
	generator     *generator.Generator // Domain: Go code generator
}

// NewApplicationService creates a new application service instance.
// It requires implementations of the input/output ports and domain services.
func NewApplicationService(
	provider LatexProvider,
	writer GoCodeWriter,
	parser *parser.Parser,
	generator *generator.Generator,
) *ApplicationService {
	return &ApplicationService{
		latexProvider: provider,
		codeWriter:    writer,
		parser:        parser,
		generator:     generator,
	}
}

// Run executes the main application logic: parse LaTeX and generate Go code.
func (s *ApplicationService) Run() error {
	// 1. Get input from the provider
	latexInput, config, err := s.latexProvider.GetLatexInput()
	if err != nil {
		return fmt.Errorf("failed to get latex input: %w", err)
	}

	// 2. Parse the LaTeX string using the domain parser
	internalAST, err := s.parser.Parse(latexInput)
	if err != nil {
		return fmt.Errorf("failed to parse latex: %w", err)
	}

	// 3. Generate Go code using the domain generator
	goCode, err := s.generator.Generate(internalAST, config.PackageName, config.FuncName)
	if err != nil {
		return fmt.Errorf("failed to generate go code: %w", err)
	}

	// 4. Write the output using the code writer
	err = s.codeWriter.WriteGoCode(goCode)
	if err != nil {
		return fmt.Errorf("failed to write go code: %w", err)
	}

	fmt.Println("Successfully generated Go code.") // Add success message
	return nil
}
