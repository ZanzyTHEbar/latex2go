package main

import (
	"fmt"
	"log" // Use log for fatal errors
	"os"

	// Application core & domain
	"github.com/ZanzyTHEbar/latex2go/internal/app"
	"github.com/ZanzyTHEbar/latex2go/internal/domain/generator"
	"github.com/ZanzyTHEbar/latex2go/internal/domain/parser"

	// Adapters
	"github.com/ZanzyTHEbar/latex2go/internal/adapters/cli"
	"github.com/ZanzyTHEbar/latex2go/internal/adapters/output"

	"github.com/spf13/cobra"
)

// Flags are now handled directly by Cobra within init() and Run()
// var (
// 	inputFile  string
// 	outputFile string
// 	packageName string
// 	funcName    string
// )

var rootCmd = &cobra.Command{
	Use:   "latex2go",
	Short: "latex2go converts LaTeX math equations to Go code",
	Long: `latex2go is a CLI tool that takes a LaTeX mathematical equation
as input and generates equivalent Go code.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Retrieve flag values needed for adapter creation
		outputFilePath, _ := cmd.Flags().GetString("output") // Error checked by Cobra

		// --- Dependency Injection ---
		// 1. Instantiate Domain Services
		latexParser := parser.NewParser()
		codeGenerator := generator.NewGenerator()

		// 2. Instantiate Adapters
		// Input adapter uses the command itself to access flags
		inputAdapter := cli.NewAdapter(cmd)
		// Output adapter uses the factory based on the output path flag
		outputAdapter := output.NewWriterAdapter(outputFilePath)

		// 3. Instantiate Application Service
		appService := app.NewApplicationService(inputAdapter, outputAdapter, latexParser, codeGenerator)

		// --- Execute Application Logic ---
		err := appService.Run()
		if err != nil {
			// Log the error to stderr and exit
			log.Fatalf("Error: %v\n", err)
			// os.Exit(1) // log.Fatalf exits with status 1
		}
	},
}

func init() {
	// Define flags using Cobra's recommended practice (accessing via cmd.Flags() in Run)
	rootCmd.Flags().StringP("input", "i", "", "LaTeX equation string (required)")
	rootCmd.Flags().StringP("output", "o", "", "Output Go file path (default: stdout)")
	rootCmd.Flags().String("package", "main", "Go package name for the generated file")
	rootCmd.Flags().String("func-name", "calculate", "Function name in the generated Go code")

	// Mark input as required
	if err := rootCmd.MarkFlagRequired("input"); err != nil {
		// This error handling is for programming errors during setup
		fmt.Fprintf(os.Stderr, "Error marking flag required: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra handles reporting the error to stderr here
		os.Exit(1)
	}
}
