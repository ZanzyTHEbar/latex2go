package app_test

import (
	"errors"
	"testing"

	"github.com/ZanzyTHEbar/latex2go/internal/app"
	app_mocks "github.com/ZanzyTHEbar/latex2go/internal/app/mocks"
	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
	gen_mocks "github.com/ZanzyTHEbar/latex2go/internal/domain/generator/mocks"
	parser_mocks "github.com/ZanzyTHEbar/latex2go/internal/domain/parser/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationService_Run_Success(t *testing.T) {
	// Arrange
	mockProvider := app_mocks.NewMockLatexProvider(t)
	mockWriter := app_mocks.NewMockGoCodeWriter(t)
	mockParser := parser_mocks.NewMockParser(t)
	mockGenerator := gen_mocks.NewMockGenerator(t)

	inputLatex := "a + b"
	inputConfig := app.Config{
		OutputFile:  "output.go",
		PackageName: "testpkg",
		FuncName:    "testFunc",
	}
	mockAST := &ast.BinaryExpr{ // A simple placeholder AST
		Op:    "+",
		Left:  &ast.Variable{Name: "a"},
		Right: &ast.Variable{Name: "b"},
	}
	expectedGoCode := "package testpkg\n\nimport \"math\"\n\nfunc testFunc(a float64, b float64) float64 {\n\treturn a + b\n}"

	// Setup mock expectations
	mockProvider.On("GetLatexInput").Return(inputLatex, inputConfig, nil).Once()
	mockParser.On("Parse", inputLatex).Return(mockAST, nil).Once()
	mockGenerator.On("Generate", mockAST, inputConfig.PackageName, inputConfig.FuncName).Return(expectedGoCode, nil).Once()
	mockWriter.On("WriteGoCode", expectedGoCode).Return(nil).Once()

	// Instantiate the service with mocks
	// Note: We pass the concrete mock types which satisfy the interfaces
	service := app.NewApplicationService(mockProvider, mockWriter, mockParser, mockGenerator)

	// Act
	err := service.Run()

	// Assert
	require.NoError(t, err)
	// AssertExpectations(t) is called automatically by testify's cleanup
}

func TestApplicationService_Run_GetInputError(t *testing.T) {
	// Arrange
	mockProvider := app_mocks.NewMockLatexProvider(t)
	mockWriter := app_mocks.NewMockGoCodeWriter(t)
	mockParser := parser_mocks.NewMockParser(t)
	mockGenerator := gen_mocks.NewMockGenerator(t)

	expectedError := errors.New("failed to get input")
	mockProvider.On("GetLatexInput").Return("", app.Config{}, expectedError).Once()

	service := app.NewApplicationService(mockProvider, mockWriter, mockParser, mockGenerator)

	// Act
	err := service.Run()

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to get latex input")
	assert.ErrorIs(t, err, expectedError)
}

func TestApplicationService_Run_ParseError(t *testing.T) {
	// Arrange
	mockProvider := app_mocks.NewMockLatexProvider(t)
	mockWriter := app_mocks.NewMockGoCodeWriter(t)
	mockParser := parser_mocks.NewMockParser(t)
	mockGenerator := gen_mocks.NewMockGenerator(t)

	inputLatex := "invalid latex"
	inputConfig := app.Config{PackageName: "p", FuncName: "f"}
	expectedError := errors.New("parsing failed")

	mockProvider.On("GetLatexInput").Return(inputLatex, inputConfig, nil).Once()
	mockParser.On("Parse", inputLatex).Return(nil, expectedError).Once()

	service := app.NewApplicationService(mockProvider, mockWriter, mockParser, mockGenerator)

	// Act
	err := service.Run()

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to parse latex")
	assert.ErrorIs(t, err, expectedError)
}

func TestApplicationService_Run_GenerateError(t *testing.T) {
	// Arrange
	mockProvider := app_mocks.NewMockLatexProvider(t)
	mockWriter := app_mocks.NewMockGoCodeWriter(t)
	mockParser := parser_mocks.NewMockParser(t)
	mockGenerator := gen_mocks.NewMockGenerator(t)

	inputLatex := "a^b"
	inputConfig := app.Config{PackageName: "p", FuncName: "f"}
	mockAST := &ast.BinaryExpr{Op: "^", Left: &ast.Variable{Name: "a"}, Right: &ast.Variable{Name: "b"}}
	expectedError := errors.New("generation failed")

	mockProvider.On("GetLatexInput").Return(inputLatex, inputConfig, nil).Once()
	mockParser.On("Parse", inputLatex).Return(mockAST, nil).Once()
	mockGenerator.On("Generate", mockAST, inputConfig.PackageName, inputConfig.FuncName).Return("", expectedError).Once()

	service := app.NewApplicationService(mockProvider, mockWriter, mockParser, mockGenerator)

	// Act
	err := service.Run()

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to generate go code")
	assert.ErrorIs(t, err, expectedError)
}

func TestApplicationService_Run_WriteError(t *testing.T) {
	// Arrange
	mockProvider := app_mocks.NewMockLatexProvider(t)
	mockWriter := app_mocks.NewMockGoCodeWriter(t)
	mockParser := parser_mocks.NewMockParser(t)
	mockGenerator := gen_mocks.NewMockGenerator(t)

	inputLatex := "x"
	inputConfig := app.Config{PackageName: "p", FuncName: "f"}
	mockAST := &ast.Variable{Name: "x"}
	generatedCode := "package p..."
	expectedError := errors.New("write failed")

	mockProvider.On("GetLatexInput").Return(inputLatex, inputConfig, nil).Once()
	mockParser.On("Parse", inputLatex).Return(mockAST, nil).Once()
	mockGenerator.On("Generate", mockAST, inputConfig.PackageName, inputConfig.FuncName).Return(generatedCode, nil).Once()
	mockWriter.On("WriteGoCode", generatedCode).Return(expectedError).Once()

	service := app.NewApplicationService(mockProvider, mockWriter, mockParser, mockGenerator)

	// Act
	err := service.Run()

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to write go code")
	assert.ErrorIs(t, err, expectedError)
}
