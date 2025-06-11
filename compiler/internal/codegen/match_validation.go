package codegen

import (
	"sort"

	"github.com/christianfindlay/osprey/internal/ast"
)

// validateMatchExpression validates match expressions for exhaustiveness and unknown variants.
func (g *LLVMGenerator) validateMatchExpression(matchExpr *ast.MatchExpression) error {
	// Get the discriminant type if it's a known union type
	var unionType *ast.TypeDeclaration
	if ident, ok := matchExpr.Expression.(*ast.Identifier); ok {
		if varType, exists := g.variableTypes[ident.Name]; exists {
			if typeDecl, exists := g.typeDeclarations[varType]; exists {
				unionType = typeDecl
			}
		}
	}

	if unionType == nil {
		// If we can't determine the union type, skip validation for now
		return nil
	}

	// Validate that all patterns are known variants
	if err := g.validatePatternVariants(matchExpr.Arms, unionType); err != nil {
		return err
	}

	// Validate exhaustiveness
	if err := g.validateExhaustiveness(matchExpr.Arms, unionType); err != nil {
		return err
	}

	return nil
}

// validatePatternVariants ensures all patterns in the match arms are valid variants.
func (g *LLVMGenerator) validatePatternVariants(arms []ast.MatchArm, unionType *ast.TypeDeclaration) error {
	for _, arm := range arms {
		if arm.Pattern.Constructor == "_" || arm.Pattern.Constructor == UnknownPattern {
			continue // Wildcard patterns are always valid
		}

		// Check if the pattern is a valid variant
		found := false
		for _, variant := range unionType.Variants {
			if arm.Pattern.Constructor == variant.Name {
				found = true

				break
			}
		}

		if !found {
			return WrapMatchUnknownVariantType(arm.Pattern.Constructor, unionType.Name)
		}
	}

	return nil
}

// validateExhaustiveness ensures all variants of the union type are covered.
func (g *LLVMGenerator) validateExhaustiveness(arms []ast.MatchArm, unionType *ast.TypeDeclaration) error {
	// Collect all covered patterns
	coveredPatterns := make(map[string]bool)
	hasWildcard := false

	for _, arm := range arms {
		if arm.Pattern.Constructor == "_" || arm.Pattern.Constructor == UnknownPattern {
			hasWildcard = true
		} else {
			coveredPatterns[arm.Pattern.Constructor] = true
		}
	}

	// If there's a wildcard, match is exhaustive
	if hasWildcard {
		return nil
	}

	// Check if all variants are covered
	var missingPatterns []string
	for _, variant := range unionType.Variants {
		if !coveredPatterns[variant.Name] {
			missingPatterns = append(missingPatterns, variant.Name)
		}
	}

	if len(missingPatterns) > 0 {
		sort.Strings(missingPatterns)

		return WrapMatchNotExhaustive(missingPatterns)
	}

	return nil
}
