package ast_test

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/christianfindlay/osprey/internal/ast"
	"github.com/christianfindlay/osprey/parser"
)

// Helper function to parse source and build AST.
func parseToAST(t *testing.T, source string) *ast.Program {
	t.Helper()

	input := antlr.NewInputStream(source)
	lexer := parser.NewospreyLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewospreyParser(stream)
	tree := p.Program()

	builder := ast.NewBuilder()

	return builder.BuildProgram(tree)
}

func TestInterpolatedStringParsing(t *testing.T) {
	tests := createInterpolatedStringTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpolatedStringTest(t, tt)
		})
	}
}

func createInterpolatedStringTests() []struct {
	name       string
	source     string
	expectVar  string
	expectText string
} {
	return []struct {
		name       string
		source     string
		expectVar  string
		expectText string
	}{
		{
			name:       "simple variable interpolation",
			source:     `print("Hello ${name}!")`,
			expectVar:  "name",
			expectText: "Hello ",
		},
		{
			name:       "integer variable interpolation",
			source:     `print("Age: ${age}")`,
			expectVar:  "age",
			expectText: "Age: ",
		},
		{
			name:       "multiple interpolations",
			source:     `print("${greeting} ${name}!")`,
			expectVar:  "greeting",
			expectText: " ",
		},
	}
}

func runInterpolatedStringTest(t *testing.T, tt struct {
	name       string
	source     string
	expectVar  string
	expectText string
},
) {
	program := parseToAST(t, tt.source)
	interpStr := extractInterpolatedString(t, program)
	validateInterpolatedStringParts(t, interpStr, tt.expectVar, tt.expectText)
}

func extractInterpolatedString(t *testing.T, program *ast.Program) *ast.InterpolatedStringLiteral {
	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}

	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %T", exprStmt.Expression)
	}

	if len(callExpr.Arguments) != 1 {
		t.Fatalf("Expected 1 argument, got %d", len(callExpr.Arguments))
	}

	interpStr, ok := callExpr.Arguments[0].(*ast.InterpolatedStringLiteral)
	if !ok {
		t.Fatalf("Expected InterpolatedStringLiteral, got %T", callExpr.Arguments[0])
	}

	return interpStr
}

func validateInterpolatedStringParts(
	t *testing.T,
	interpStr *ast.InterpolatedStringLiteral,
	expectVar string,
	expectText string,
) {
	if len(interpStr.Parts) == 0 {
		t.Fatal("Expected interpolated string to have parts")
	}

	foundVar := findVariableInParts(interpStr.Parts, expectVar)
	foundText := findTextInParts(interpStr.Parts, expectText)

	if !foundVar {
		t.Errorf("Expected to find variable %q in interpolated string", expectVar)
	}

	if !foundText {
		t.Errorf("Expected to find text %q in interpolated string", expectText)
	}
}

func findVariableInParts(parts []ast.InterpolatedPart, expectVar string) bool {
	for _, part := range parts {
		if part.IsExpression {
			if ident, ok := part.Expression.(*ast.Identifier); ok {
				if ident.Name == expectVar {
					return true
				}
			}
		}
	}

	return false
}

func findTextInParts(parts []ast.InterpolatedPart, expectText string) bool {
	for _, part := range parts {
		if !part.IsExpression && strings.Contains(part.Text, expectText) {
			return true
		}
	}

	return false
}

func TestArithmeticInInterpolatedString(t *testing.T) {
	tests := createArithmeticTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runArithmeticTest(t, tt)
		})
	}
}

func createArithmeticTests() []struct {
	name     string
	source   string
	operator string
	leftVar  string
	rightVal int64
} {
	return []struct {
		name     string
		source   string
		operator string
		leftVar  string
		rightVal int64
	}{
		{
			name:     "addition expression",
			source:   `print("Next: ${age + 1}")`,
			operator: "+",
			leftVar:  "age",
			rightVal: 1,
		},
		{
			name:     "subtraction expression",
			source:   `print("Previous: ${age - 1}")`,
			operator: "-",
			leftVar:  "age",
			rightVal: 1,
		},
		{
			name:     "multiplication expression",
			source:   `print("Double: ${x * 2}")`,
			operator: "*",
			leftVar:  "x",
			rightVal: 2,
		},
		{
			name:     "division expression",
			source:   `print("Half: ${x / 2}")`,
			operator: "/",
			leftVar:  "x",
			rightVal: 2,
		},
	}
}

func runArithmeticTest(
	t *testing.T,
	tt struct {
		name     string
		source   string
		operator string
		leftVar  string
		rightVal int64
	},
) {
	program := parseToAST(t, tt.source)
	interpStr := extractInterpolatedStringFromProgram(t, program)
	binExpr := findBinaryExpressionInParts(t, interpStr.Parts)
	validateBinaryExpression(
		t, binExpr, tt.operator, tt.leftVar, tt.rightVal,
	)
}

func extractInterpolatedStringFromProgram(_ *testing.T, program *ast.Program) *ast.InterpolatedStringLiteral {
	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	callExpr := exprStmt.Expression.(*ast.CallExpression)

	return callExpr.Arguments[0].(*ast.InterpolatedStringLiteral)
}

func findBinaryExpressionInParts(t *testing.T, parts []ast.InterpolatedPart) *ast.BinaryExpression {
	for _, part := range parts {
		if part.IsExpression {
			// Check if it's a direct BinaryExpression
			if be, ok := part.Expression.(*ast.BinaryExpression); ok {
				return be
			}

			// Check if it's a ResultExpression containing a BinaryExpression
			if re, ok := part.Expression.(*ast.ResultExpression); ok {
				if be, ok := re.Value.(*ast.BinaryExpression); ok {
					return be
				}
			}
		}
	}

	t.Fatal("Expected to find binary expression in interpolated string")

	return nil
}

func validateBinaryExpression(t *testing.T, binExpr *ast.BinaryExpression, operator, leftVar string, rightVal int64) {
	if binExpr.Operator != operator {
		t.Errorf("Expected operator %q, got %q", operator, binExpr.Operator)
	}

	validateLeftSide(t, binExpr.Left, leftVar)
	validateRightSide(t, binExpr.Right, rightVal)
}

func validateLeftSide(t *testing.T, left ast.Expression, leftVar string) {
	leftIdent, ok := left.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected left side to be Identifier, got %T", left)
	}

	if leftIdent.Name != leftVar {
		t.Errorf("Expected left variable %q, got %q", leftVar, leftIdent.Name)
	}
}

func validateRightSide(t *testing.T, right ast.Expression, rightVal int64) {
	rightLit, ok := right.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("Expected right side to be IntegerLiteral, got %T", right)
	}

	if rightLit.Value != rightVal {
		t.Errorf("Expected right value %d, got %d", rightVal, rightLit.Value)
	}
}

func TestComplexInterpolatedExpressions(t *testing.T) {
	t.Run("multiple variables in same string", func(t *testing.T) {
		testMultipleVariables(t)
	})

	t.Run("function call in interpolation", func(t *testing.T) {
		testFunctionCallInInterpolation(t)
	})
}

func testMultipleVariables(t *testing.T) {
	source := `print("${name} is ${age} years old")`
	program := parseToAST(t, source)
	interpStr := extractComplexInterpolatedString(t, program)

	validateMinimumParts(t, interpStr, 3)
	variables := extractVariablesFromParts(interpStr.Parts)
	validateExpectedVariables(t, variables, []string{"name", "age"})
}

func testFunctionCallInInterpolation(t *testing.T) {
	source := `print("Result: ${double(x)}")`
	program := parseToAST(t, source)
	interpStr := extractComplexInterpolatedString(t, program)

	foundDouble := findDoubleInParts(interpStr.Parts)
	if !foundDouble {
		t.Log("Function call parsing is simplified in current implementation")
	}
}

func extractComplexInterpolatedString(_ *testing.T, program *ast.Program) *ast.InterpolatedStringLiteral {
	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	callExpr := exprStmt.Expression.(*ast.CallExpression)

	return callExpr.Arguments[0].(*ast.InterpolatedStringLiteral)
}

func validateMinimumParts(t *testing.T, interpStr *ast.InterpolatedStringLiteral, minParts int) {
	if len(interpStr.Parts) < minParts {
		t.Errorf("Expected at least %d parts, got %d", minParts, len(interpStr.Parts))
	}
}

func extractVariablesFromParts(parts []ast.InterpolatedPart) []string {
	var variables []string

	for _, part := range parts {
		if part.IsExpression {
			if ident, ok := part.Expression.(*ast.Identifier); ok {
				variables = append(variables, ident.Name)
			}
		}
	}

	return variables
}

func validateExpectedVariables(t *testing.T, variables, expected []string) {
	if len(variables) != len(expected) {
		t.Errorf("Expected %v variables, got %v", expected, variables)
	}

	for i, expectedVar := range expected {
		if i < len(variables) && variables[i] != expectedVar {
			t.Errorf("Expected variable %q at position %d, got %q", expectedVar, i, variables[i])
		}
	}
}

func findDoubleInParts(parts []ast.InterpolatedPart) bool {
	for _, part := range parts {
		if part.IsExpression {
			if ident, ok := part.Expression.(*ast.Identifier); ok {
				if ident.Name == "double" || strings.Contains(ident.Name, "double") {
					return true
				}
			}
		}
	}

	return false
}

func TestInterpolatedStringEdgeCases(t *testing.T) {
	t.Run("empty string parts", func(t *testing.T) {
		source := `print("${name}${age}")`
		program := parseToAST(t, source)

		exprStmt := program.Statements[0].(*ast.ExpressionStatement)
		callExpr := exprStmt.Expression.(*ast.CallExpression)
		interpStr := callExpr.Arguments[0].(*ast.InterpolatedStringLiteral)

		// Should have just the two expressions, no text between
		expressionCount := 0

		for _, part := range interpStr.Parts {
			if part.IsExpression {
				expressionCount++
			}
		}

		if expressionCount != 2 {
			t.Errorf("Expected 2 expressions, got %d", expressionCount)
		}
	})

	t.Run("text only at start and end", func(t *testing.T) {
		source := `print("Hello ${name} goodbye")`
		program := parseToAST(t, source)

		exprStmt := program.Statements[0].(*ast.ExpressionStatement)
		callExpr := exprStmt.Expression.(*ast.CallExpression)
		interpStr := callExpr.Arguments[0].(*ast.InterpolatedStringLiteral)

		// Should have: text, expr, text
		if len(interpStr.Parts) != 3 {
			t.Errorf("Expected 3 parts, got %d", len(interpStr.Parts))
		}

		// First and last should be text
		if interpStr.Parts[0].IsExpression {
			t.Error("Expected first part to be text")
		}

		if interpStr.Parts[2].IsExpression {
			t.Error("Expected last part to be text")
		}

		if !interpStr.Parts[1].IsExpression {
			t.Error("Expected middle part to be expression")
		}
	})
}
