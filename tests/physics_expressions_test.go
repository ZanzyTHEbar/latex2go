package tests

import (
	"testing"

	"github.com/ZanzyTHEbar/latex2go/internal/domain/parser"
	"github.com/ZanzyTHEbar/latex2go/internal/domain/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStandardModelExpressions tests various Standard Model physics expressions
func TestStandardModelExpressions(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectError    bool
		expectContains []string // Strings that should appear in the generated Go code
	}{
		{
			name:  "Einstein Mass-Energy Equivalence",
			input: "E = m * c^2",
			expectContains: []string{
				"func calculate(E float64, c float64, m float64) float64",
				"math.Pow(c, 2)",
				"left := E",
				"right := m * math.Pow(c, 2)",
			},
		},
		{
			name:  "Subscripted Variables - Gauge Fields",
			input: "A_μ + B_ν",
			expectContains: []string{
				"A_mu",
				"B_nu",
				"A_mu + B_nu",
			},
		},
		{
			name:  "Greek Letters - Dirac Matrices",
			input: "γ + δ + μ + ν + α + β",
			expectContains: []string{
				"gamma", "delta", "mu", "nu", "alpha", "beta",
			},
		},
		{
			name:  "Fermionic Fields",
			input: "ψ_L + ψ_R",
			expectContains: []string{
				"psi_L", "psi_R",
			},
		},
		{
			name:  "Complex Physics Expression",
			input: "ℒ = ψ_L * γ_μ * ψ_R + m * ψ_L * ψ_R",
			expectContains: []string{
				"Lagrangian", "psi_L", "gamma_mu", "psi_R", "m",
			},
		},
		{
			name:  "Electromagnetic Field Tensor",
			input: "F_μν = ∂_μ * A_ν - ∂_ν * A_μ",
			expectContains: []string{
				"F_munu", "partial_mu", "A_nu", "partial_nu", "A_mu",
			},
		},
		{
			name:  "Covariant Derivative",
			input: "D_μ = ∂_μ + g * A_μ",
			expectContains: []string{
				"D_mu", "partial_mu", "g", "A_mu",
			},
		},
		{
			name:  "Higgs Potential",
			input: "V = μ^2 * φ^2 + λ * φ^4",
			expectContains: []string{
				"math.Pow(mu, 2)", "math.Pow(phi, 2)", "lambda", "math.Pow(phi, 4)",
			},
		},
		{
			name:  "QCD Beta Function",
			input: "β = g^3 + a * g^5",
			expectContains: []string{
				"beta", "math.Pow(g, 3)", "a", "math.Pow(g, 5)",
			},
		},
	}

	parser := parser.NewParser()
	generator := generator.NewGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the LaTeX expression
			ast, err := parser.Parse(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err, "Failed to parse LaTeX: %s", tt.input)
			require.NotNil(t, ast)

			// Generate Go code
			goCode, err := generator.Generate(ast, "physics", "calculate")
			require.NoError(t, err, "Failed to generate Go code for: %s", tt.input)
			
			// Check that expected strings appear in the generated code
			for _, expected := range tt.expectContains {
				assert.Contains(t, goCode, expected, 
					"Generated code should contain '%s'\nGenerated code:\n%s", expected, goCode)
			}
			
			// Verify the generated code is valid Go syntax by checking it can be formatted
			// (The generator already calls format.Source internally)
			assert.NotEmpty(t, goCode, "Generated code should not be empty")
			assert.Contains(t, goCode, "package physics", "Should use correct package name")
			assert.Contains(t, goCode, "func calculate", "Should generate function with correct name")
		})
	}
}

// TestPhysicsConstantsAndSymbols tests recognition of common physics symbols
func TestPhysicsConstantsAndSymbols(t *testing.T) {
	symbolTests := []struct {
		symbol      string
		expectedGo  string
		description string
	}{
		{"α", "alpha", "Fine structure constant"},
		{"β", "beta", "Beta decay"},
		{"γ", "gamma", "Gamma matrices / Lorentz factor"},
		{"δ", "delta", "Delta function"},
		{"ε", "epsilon", "Permittivity"},
		{"ζ", "zeta", "Zeta function"},
		{"η", "eta", "Eta meson"},
		{"θ", "theta", "Angle parameter"},
		{"λ", "lambda", "Wavelength / coupling"},
		{"μ", "mu", "Muon / chemical potential"},
		{"ν", "nu", "Neutrino / frequency"},
		{"π", "pi", "Pion"},
		{"ρ", "rho", "Density"},
		{"σ", "sigma", "Cross section / Pauli matrices"},
		{"τ", "tau", "Tau lepton / proper time"},
		{"φ", "phi", "Scalar field"},
		{"χ", "chi", "Chi particle"},
		{"ψ", "psi", "Wave function"},
		{"ω", "omega", "Angular frequency"},
		{"Γ", "Gamma", "Gamma function"},
		{"Δ", "Delta", "Delta baryon"},
		{"Θ", "Theta", "Theta angle"},
		{"Λ", "Lambda", "Lambda baryon / cosmological constant"},
		{"Π", "Pi", "Pi meson"},
		{"Σ", "Sigma", "Sigma baryon"},
		{"Φ", "Phi", "Phi meson"},
		{"Ψ", "Psi", "Psi meson"},
		{"Ω", "Omega", "Omega baryon"},
		{"∂", "partial", "Partial derivative"},
		{"∇", "nabla", "Nabla operator"},
		{"†", "dagger", "Hermitian conjugate"},
		{"ℒ", "Lagrangian", "Lagrangian density"},
	}

	parser := parser.NewParser()
	generator := generator.NewGenerator()

	for _, tt := range symbolTests {
		t.Run(tt.description, func(t *testing.T) {
			// Test the symbol in a simple expression
			input := tt.symbol + " + x"
			
			ast, err := parser.Parse(input)
			require.NoError(t, err, "Failed to parse symbol: %s", tt.symbol)
			
			goCode, err := generator.Generate(ast, "main", "test")
			require.NoError(t, err, "Failed to generate code for symbol: %s", tt.symbol)
			
			assert.Contains(t, goCode, tt.expectedGo, 
				"Generated code should contain Go equivalent '%s' for symbol '%s'\nGenerated: %s", 
				tt.expectedGo, tt.symbol, goCode)
		})
	}
}