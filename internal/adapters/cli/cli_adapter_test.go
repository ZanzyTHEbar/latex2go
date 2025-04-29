package cli_test

import (
	"testing"

	"github.com/ZanzyTHEbar/latex2go/internal/adapters/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCliAdapter_GetLatexInput_Success(t *testing.T) {
	// Arrange
	cmd := &cobra.Command{}
	cmd.Flags().StringP("input", "i", "", "LaTeX equation string")
	cmd.Flags().StringP("output", "o", "", "Output Go file path")
	cmd.Flags().String("package", "main", "Go package name")
	cmd.Flags().String("func-name", "calculate", "Function name")

	// Set flag values for the test
	expectedLatex := "x^2 + y^2"
	expectedOutput := "calc.go"
	expectedPackage := "mathops"
	expectedFunc := "compute"

	cmd.Flags().Set("input", expectedLatex)
	cmd.Flags().Set("output", expectedOutput)
	cmd.Flags().Set("package", expectedPackage)
	cmd.Flags().Set("func-name", expectedFunc)

	adapter := cli.NewAdapter(cmd)

	// Act
	latex, config, err := adapter.GetLatexInput()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedLatex, latex)
	assert.Equal(t, expectedOutput, config.OutputFile)
	assert.Equal(t, expectedPackage, config.PackageName)
	assert.Equal(t, expectedFunc, config.FuncName)
}

func TestCliAdapter_GetLatexInput_MissingInput(t *testing.T) {
	// Arrange
	cmd := &cobra.Command{}
	cmd.Flags().StringP("input", "i", "", "LaTeX equation string")
	cmd.Flags().StringP("output", "o", "", "Output Go file path")
	cmd.Flags().String("package", "main", "Go package name")
	cmd.Flags().String("func-name", "calculate", "Function name")

	// Input flag is deliberately not set

	adapter := cli.NewAdapter(cmd)

	// Act
	_, _, err := adapter.GetLatexInput()

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "input LaTeX string cannot be empty")
}

func TestCliAdapter_NewAdapter_PanicMissingFlags(t *testing.T) {
	// Arrange
	cmd := &cobra.Command{}
	// Deliberately omit defining flags

	// Act & Assert
	assert.PanicsWithValue(t,
		"CLI Adapter requires command with 'input', 'output', 'package', and 'func-name' flags defined",
		func() { cli.NewAdapter(cmd) },
		"Should panic if flags are missing",
	)
}
