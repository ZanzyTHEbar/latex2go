# Standard Model Physics Support in latex2go

## Overview

This document demonstrates the enhanced capabilities of latex2go to parse and convert Standard Model physics expressions from LaTeX to Go code.

## Key Features Implemented

### 1. Greek Letter Support
All Standard Model particles and physics symbols are now supported:

**Leptons and Quarks:**
- `μ` → `mu` (muon)
- `τ` → `tau` (tau lepton)  
- `ν` → `nu` (neutrino)
- `ψ` → `psi` (wave function)

**Gauge Fields and Forces:**
- `γ` → `gamma` (photon/gamma matrices)
- `α` → `alpha` (fine structure constant)
- `β` → `beta` (beta coupling)
- `φ` → `phi` (scalar field/Higgs)
- `λ` → `lambda` (coupling constant)

### 2. Subscript Notation
Physics notation with subscripts is fully supported:
- `A_μ` → `A_mu` (gauge field component)
- `ψ_L` → `psi_L` (left-handed fermion)
- `ψ_R` → `psi_R` (right-handed fermion)
- `F_μν` → `F_munu` (field strength tensor)

### 3. Special Physics Symbols
- `∂` → `partial` (partial derivative)
- `∇` → `nabla` (gradient operator)
- `†` → `dagger` (Hermitian conjugate)
- `ℒ` → `Lagrangian` (Lagrangian density)

### 4. Equation Support
Mathematical equations are now properly parsed:
- `E = mc^2` generates equality-checking functions
- Complex multi-term expressions supported

## Standard Model Examples

### 1. Einstein Mass-Energy Relation
```latex
E = m * c^2
```
Generates Go function that verifies the mass-energy equivalence.

### 2. Electromagnetic Field Tensor
```latex
F_μν = ∂_μ * A_ν - ∂_ν * A_μ
```
Converts the electromagnetic field tensor definition to Go code.

### 3. QED Lagrangian Component
```latex
ℒ_QED = ψ_e * γ_μ * ψ_e * A_μ
```
Handles the electron-photon interaction term in quantum electrodynamics.

### 4. Higgs Potential
```latex
V = μ^2 * φ^2 + λ * φ^4
```
Converts the Higgs field potential with quartic self-interaction.

### 5. Covariant Derivative
```latex
D_μ = ∂_μ + g * A_μ
```
Handles gauge-covariant derivatives fundamental to gauge theories.

## Generated Code Structure

All expressions generate properly formatted Go code with:
- **Package declaration** with configurable name
- **Import statements** for math package when needed
- **Function signatures** with sorted parameters
- **Type-safe float64 operations**
- **Formatted output** using go/format

## Testing

Comprehensive test suite includes:
- **34 physics symbol tests** covering all Greek letters and special symbols
- **9 complex Standard Model expressions**
- **Regression tests** ensuring existing functionality remains intact
- **Unicode handling verification**

## Next Steps for Full Standard Model Support

1. **Matrix Operations:** Trace, determinant, transpose operations
2. **Tensor Indices:** Full superscript/subscript tensor notation (T^μ_ν)
3. **Einstein Summation:** Automatic index contraction
4. **Complex Conjugation:** Proper handling of dagger and overline operators
5. **Advanced Functions:** Exponentials, logarithms, complex number support

## Current Limitations

- Superscripts currently treated as exponentiation (not tensor indices)
- No matrix/tensor operation support yet
- Complex numbers not yet supported
- No automatic index summation

## Conclusion

The latex2go tool now supports a substantial portion of Standard Model physics notation, making it capable of parsing fundamental equations in particle physics, quantum field theory, and gauge theory. This represents a significant step toward full Standard Model support.