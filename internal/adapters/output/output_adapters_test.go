package output_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/latex2go/internal/adapters/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to capture stdout
func captureStdout(f func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}

func TestStdoutAdapter_WriteGoCode(t *testing.T) {
	// Arrange
	adapter := output.NewStdoutAdapter()
	expectedCode := "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"

	// Act
	outputStr, err := captureStdout(func() error {
		return adapter.WriteGoCode(expectedCode)
	})

	// Assert
	require.NoError(t, err)
	// fmt.Println adds a newline, so we expect the code + newline
	assert.Equal(t, expectedCode+"\n", outputStr)
}

func TestFileAdapter_WriteGoCode_NewFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir() // Creates a temporary directory cleaned up automatically
	testFilePath := filepath.Join(tempDir, "test_output.go")
	expectedCode := "package test\n\nfunc run() {}"

	adapter := output.NewFileAdapter(testFilePath)

	// Act
	err := adapter.WriteGoCode(expectedCode)

	// Assert
	require.NoError(t, err)

	// Verify file content
	contentBytes, readErr := os.ReadFile(testFilePath)
	require.NoError(t, readErr)
	assert.Equal(t, expectedCode, string(contentBytes))
}

func TestFileAdapter_WriteGoCode_OverwriteFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "test_output_overwrite.go")
	initialContent := "initial content"
	expectedCode := "package overwrite\n\nfunc main() {}"

	// Create initial file
	require.NoError(t, os.WriteFile(testFilePath, []byte(initialContent), 0644))

	adapter := output.NewFileAdapter(testFilePath)

	// Act
	err := adapter.WriteGoCode(expectedCode)

	// Assert
	require.NoError(t, err)

	// Verify file content is overwritten
	contentBytes, readErr := os.ReadFile(testFilePath)
	require.NoError(t, readErr)
	assert.Equal(t, expectedCode, string(contentBytes))
}

func TestFileAdapter_WriteGoCode_InvalidPath(t *testing.T) {
	// Arrange
	// Use a path that likely cannot be written to (e.g., root directory without permissions,
	// or a path containing invalid characters - though os.WriteFile might handle some)
	// Let's try writing to a directory as if it were a file.
	tempDir := t.TempDir()
	adapter := output.NewFileAdapter(tempDir) // Path is a directory
	expectedCode := "package fail"

	// Act
	err := adapter.WriteGoCode(expectedCode)

	// Assert
	require.Error(t, err)
	// The exact error might vary by OS, but it should indicate a write failure
	assert.ErrorContains(t, err, "failed to write code to file")
	fmt.Println("Expected error writing to directory:", err) // Log for debugging if needed
}

func TestNewFileAdapter_PanicEmptyPath(t *testing.T) {
	// Arrange, Act & Assert
	assert.PanicsWithValue(t,
		"FileAdapter requires a non-empty file path",
		func() { output.NewFileAdapter("") },
		"Should panic if file path is empty",
	)
}

func TestNewWriterAdapter_Factory(t *testing.T) {
	t.Run("Empty Path returns StdoutAdapter", func(t *testing.T) {
		adapter := output.NewWriterAdapter("")
		assert.IsType(t, &output.StdoutAdapter{}, adapter)
	})

	t.Run("Non-Empty Path returns FileAdapter", func(t *testing.T) {
		adapter := output.NewWriterAdapter("some/path.go")
		assert.IsType(t, &output.FileAdapter{}, adapter)
	})
}
