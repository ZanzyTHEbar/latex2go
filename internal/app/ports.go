package app

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
