# latex2go

`latex2go` is a command-line tool written in Go that converts mathematical equations written in LaTeX format into equivalent Go code.

## Features

*   Parses LaTeX mathematical expressions.
*   Generates corresponding Go code representing the calculation.
*   Supports outputting generated code to standard output or a file.
*   Configurable package name and function name for the generated code.

## Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/ZanzyTHEbar/latex2go.git
    cd latex2go
    ```
2.  **Build the binary:**
    You can use the provided Makefile:
    ```bash
    make build
    ```
    This will create the executable `main` (or a platform-specific name if you modify the Makefile) in the project root.

    Alternatively, build directly using Go:
    ```bash
    go build -o latex2go ./cmd/latex2go/main.go
    ```

## Usage

Run the tool using the built executable:

```bash
./latex2go -i "your_latex_equation" [flags]
```

**Required Flag:**

*   `-i`, `--input`: The LaTeX equation string to convert.

**Optional Flags:**

*   `-o`, `--output`: Path to the output Go file. If not specified, the generated code will be printed to standard output.
*   `--package`: The package name for the generated Go code (default: `main`).
*   `--func-name`: The function name in the generated Go code (default: `calculate`).

**Example:**

```bash
# Output to stdout
./latex2go -i "x^2 + y^2"

# Output to a file named 'calculation.go' with package 'mathops' and function 'compute'
./latex2go -i "a / (b + c)" -o calculation.go --package mathops --func-name compute
```

## Development

The project uses a standard Go project structure and a `Makefile` for common development tasks:

*   `make build`: Build the application binary.
*   `make test`: Run unit tests.
*   `make lint`: Run linters (`golangci-lint`) and `go vet`.
*   `make tidy`: Tidy Go module dependencies.
*   `make run`: Build and run the application (currently configured for a server, might need adjustment for CLI).
*   `make docker-build`: Build a Docker image for the application.
*   `make docker-run`: Run the application using Docker Compose.
*   `make docker-down`: Stop the application running via Docker Compose.

## Architecture

The project follows a clean architecture pattern:

*   **`cmd/`**: Contains the main application entry point (`latex2go`).
*   **`internal/`**: Contains the core logic, separated into:
    *   **`adapters/`**: Adapters for interacting with external systems (CLI, file output).
    *   **`app/`**: Application service orchestrating the use cases.
    *   **`domain/`**: Core domain logic (parsing LaTeX, generating Go code).
*   **`pkg/`**: (If used) Shared libraries intended for external use.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues.

## License

(Specify your license here, e.g., MIT License)
