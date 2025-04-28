package output

import (
	"fmt"
	"os"

	"github.com/ZanzyTHEbar/latex2go/internal/app" // For app.GoCodeWriter
)

// --- Stdout Adapter ---

// StdoutAdapter implements the app.GoCodeWriter interface for stdout.
type StdoutAdapter struct{}

// NewStdoutAdapter creates a new adapter for writing to standard output.
func NewStdoutAdapter() *StdoutAdapter {
	return &StdoutAdapter{}
}

// WriteGoCode prints the generated Go code string to standard output.
func (a *StdoutAdapter) WriteGoCode(code string) error {
	_, err := fmt.Println(code) // fmt.Println writes to os.Stdout
	if err != nil {
		return fmt.Errorf("failed to write code to stdout: %w", err)
	}
	return nil
}

// --- File Adapter ---

// FileAdapter implements the app.GoCodeWriter interface for file output.
type FileAdapter struct {
	filePath string
}

// NewFileAdapter creates a new adapter for writing to a specific file.
func NewFileAdapter(filePath string) *FileAdapter {
	if filePath == "" {
		// This should ideally be prevented by logic choosing the adapter,
		// but added as a safeguard.
		panic("FileAdapter requires a non-empty file path")
	}
	return &FileAdapter{filePath: filePath}
}

// WriteGoCode writes the generated Go code string to the specified file.
// It will overwrite the file if it exists.
func (a *FileAdapter) WriteGoCode(code string) error {
	// Use os.WriteFile which handles creating/truncating the file.
	// Use 0644 permissions as a standard default for new files.
	err := os.WriteFile(a.filePath, []byte(code), 0644)
	if err != nil {
		return fmt.Errorf("failed to write code to file '%s': %w", a.filePath, err)
	}
	return nil
}

// --- Factory Function ---

// NewWriterAdapter creates the appropriate GoCodeWriter based on the output file path.
// If outputPath is empty, it returns a StdoutAdapter. Otherwise, it returns a FileAdapter.
func NewWriterAdapter(outputPath string) app.GoCodeWriter {
	if outputPath == "" {
		return NewStdoutAdapter()
	}
	return NewFileAdapter(outputPath)
}
