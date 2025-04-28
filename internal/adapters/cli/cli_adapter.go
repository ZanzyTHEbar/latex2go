package cli

import (
	"fmt"

	"github.com/ZanzyTHEbar/latex2go/internal/app" // For app.Config and app.LatexProvider
	"github.com/spf13/cobra"
)

// Adapter implements the app.LatexProvider interface using Cobra flags.
type Adapter struct {
cmd *cobra.Command
}

// NewAdapter creates a new CLI adapter instance.
func NewAdapter(cmd *cobra.Command) *Adapter {
// Ensure the necessary flags are defined on the command passed in.
// This relies on the main.go setup.
if cmd.Flag("input") == nil || cmd.Flag("output") == nil || cmd.Flag("package") == nil || cmd.Flag("func-name") == nil {
// This is a programming error check
panic("CLI Adapter requires command with 'input', 'output', 'package', and 'func-name' flags defined")
}
return &Adapter{cmd: cmd}
}

// GetLatexInput retrieves the LaTeX string and configuration from Cobra flags.
func (a *Adapter) GetLatexInput() (latex string, config app.Config, err error) {
latex, err = a.cmd.Flags().GetString("input")
if err != nil {
// This error is unlikely if the flag is correctly defined
return "", app.Config{}, fmt.Errorf("failed to get 'input' flag: %w", err)
}
if latex == "" {
// This check is technically redundant with main.go's check, but good for safety
return "", app.Config{}, fmt.Errorf("input LaTeX string cannot be empty")
}

outputFile, _ := a.cmd.Flags().GetString("output") // Error checked during flag parsing by Cobra
packageName, _ := a.cmd.Flags().GetString("package")
funcName, _ := a.cmd.Flags().GetString("func-name")

config = app.Config{
OutputFile:  outputFile,
PackageName: packageName,
FuncName:    funcName,
}

return latex, config, nil
}
