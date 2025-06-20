package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"

	"github.com/christianfindlay/osprey/internal/ast"
)

func (g *LLVMGenerator) generateCallExpression(callExpr *ast.CallExpression) (value.Value, error) {
	// Handle function calls
	if ident, ok := callExpr.Function.(*ast.Identifier); ok {
		// Handle built-in functions
		if result, err := g.handleBuiltInFunction(ident.Name, callExpr); result != nil || err != nil {
			return result, err
		}

		// Handle user-defined functions
		if fn, exists := g.functions[ident.Name]; exists {
			return g.generateUserFunctionCall(ident.Name, fn, callExpr)
		}
	}

	return nil, ErrUnsupportedCall
}

// handleBuiltInFunction handles all built-in function calls.
func (g *LLVMGenerator) handleBuiltInFunction(name string, callExpr *ast.CallExpression) (value.Value, error) {
	// Try core functions first (always available)
	if result, err := g.handleCoreFunctions(name, callExpr); result != nil || err != nil {
		return result, err
	}

	// Try HTTP functions only if allowed by security policy
	if g.security.AllowHTTP {
		if result, err := g.handleHTTPFunctions(name, callExpr); result != nil || err != nil {
			return result, err
		}
	}

	// Try WebSocket functions only if allowed by security policy
	if g.security.AllowWebSocket {
		if result, err := g.handleWebSocketFunctions(name, callExpr); result != nil || err != nil {
			return result, err
		}
	}

	// Not a built-in function or not allowed by security policy
	return nil, nil
}

// handleCoreFunctions handles core built-in functions like print, toString, etc.
func (g *LLVMGenerator) handleCoreFunctions(name string, callExpr *ast.CallExpression) (value.Value, error) {
	switch name {
	case ToStringFunc:
		return g.generateToStringCall(callExpr)
	case PrintFunc:
		return g.generatePrintCall(callExpr)
	case InputFunc:
		return g.generateInputCall(callExpr)
	case RangeFunc:
		return g.generateRangeCall(callExpr)
	case ForEachFunc:
		return g.generateForEachCall(callExpr)
	case MapFunc:
		return g.generateMapCall(callExpr)
	case FilterFunc:
		return g.generateFilterCall(callExpr)
	case FoldFunc:
		return g.generateFoldCall(callExpr)
	default:
		return nil, nil
	}
}

// handleHTTPFunctions handles HTTP-related built-in functions.
func (g *LLVMGenerator) handleHTTPFunctions(name string, callExpr *ast.CallExpression) (value.Value, error) {
	switch name {
	case HTTPCreateServerFunc:
		return g.generateHTTPCreateServerCall(callExpr)
	case HTTPListenFunc:
		return g.generateHTTPListenCall(callExpr)
	case HTTPStopServerFunc:
		return g.generateHTTPStopServerCall(callExpr)
	case HTTPCreateClientFunc:
		return g.generateHTTPCreateClientCall(callExpr)
	case HTTPGetFunc:
		return g.generateHTTPGetCall(callExpr)
	case HTTPPostFunc:
		return g.generateHTTPPostCall(callExpr)
	case HTTPPutFunc:
		return g.generateHTTPPutCall(callExpr)
	case HTTPDeleteFunc:
		return g.generateHTTPDeleteCall(callExpr)
	case HTTPRequestFunc:
		return g.generateHTTPRequestCall(callExpr)
	case HTTPCloseClientFunc:
		return g.generateHTTPCloseClientCall(callExpr)
	default:
		return nil, nil
	}
}

// handleWebSocketFunctions handles WebSocket-related built-in functions.
func (g *LLVMGenerator) handleWebSocketFunctions(name string, callExpr *ast.CallExpression) (value.Value, error) {
	switch name {
	case WebSocketConnectFunc:
		return g.generateWebSocketConnectCall(callExpr)
	case WebSocketSendFunc:
		return g.generateWebSocketSendCall(callExpr)
	case WebSocketCloseFunc:
		return g.generateWebSocketCloseCall(callExpr)
	case WebSocketCreateServerFunc:
		return g.generateWebSocketCreateServerCall(callExpr)
	case WebSocketServerListenFunc:
		return g.generateWebSocketServerListenCall(callExpr)
	case WebSocketServerBroadcastFunc:
		return g.generateWebSocketServerBroadcastCall(callExpr)
	case WebSocketStopServerFunc:
		return g.generateWebSocketStopServerCall(callExpr)
	case WebSocketKeepAlive:
		return g.generateWebSocketKeepAliveCall(callExpr)
	default:
		return nil, nil
	}
}

// generateToStringCall handles toString function calls.
func (g *LLVMGenerator) generateToStringCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != 1 {
		return nil, WrapToStringWrongArgs(len(callExpr.Arguments))
	}

	arg, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	// Check the type of the argument and handle accordingly

	return g.convertArgumentToString(callExpr.Arguments[0], arg)
}

// convertArgumentToString converts different argument types to string.
func (g *LLVMGenerator) convertArgumentToString(argExpr ast.Expression, arg value.Value) (value.Value, error) {
	switch expr := argExpr.(type) {
	case *ast.StringLiteral:
		// String -> String (identity conversion)
		return arg, nil

	case *ast.IntegerLiteral:
		// Int -> String conversion using sprintf
		return g.generateIntToString(arg)

	case *ast.BooleanLiteral:
		// Bool -> String conversion
		return g.generateBoolToString(arg)

	case *ast.Identifier:

		return g.convertIdentifierToString(expr, arg)

	case *ast.CallExpression:

		return g.convertCallExpressionToString(expr, arg)

	default:
		// For other expressions, assume int conversion
		return g.generateIntToString(arg)
	}
}

// convertIdentifierToString handles identifier to string conversion.
func (g *LLVMGenerator) convertIdentifierToString(ident *ast.Identifier, arg value.Value) (value.Value, error) {
	if varType, exists := g.variableTypes[ident.Name]; exists {
		switch varType {
		case TypeString:
			return arg, nil // Identity conversion
		case TypeInt:
			return g.generateIntToString(arg)
		case TypeBool:
			return g.generateBoolToString(arg)
		case TypeAny:
			// 'any' type is represented as i64 at LLVM level, treat as int
			return g.generateIntToString(arg)
		default:
			return nil, WrapNoToStringImpl(varType)
		}
	}
	// Default to int if type unknown

	return g.generateIntToString(arg)
}

// convertCallExpressionToString handles call expression to string conversion.
func (g *LLVMGenerator) convertCallExpressionToString(
	callArg *ast.CallExpression,
	arg value.Value,
) (value.Value, error) {
	if fnIdent, ok := callArg.Function.(*ast.Identifier); ok {
		if returnType, exists := g.functionReturnTypes[fnIdent.Name]; exists {
			switch returnType {
			case TypeString:
				return arg, nil // Identity conversion for string-returning functions
			case TypeInt:
				return g.generateIntToString(arg)
			case TypeBool:
				return g.generateBoolToString(arg)
			case TypeAny:
				// 'any' type is represented as i64 at LLVM level, treat as int
				return g.generateIntToString(arg)
			default:
				return nil, WrapNoToStringForFunc(returnType)
			}
		}
	}
	// Default to int conversion if function type unknown

	return g.generateIntToString(arg)
}

// generatePrintCall handles print function calls.
func (g *LLVMGenerator) generatePrintCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != 1 {
		return nil, WrapPrintWrongArgs(len(callExpr.Arguments))
	}

	// Auto-convert basic types to strings, reject complex types
	switch argExpr := callExpr.Arguments[0].(type) {
	case *ast.StringLiteral, *ast.InterpolatedStringLiteral:

		return g.generatePrintStringLiteral(callExpr)

	case *ast.IntegerLiteral, *ast.BinaryExpression, *ast.UnaryExpression, *ast.ResultExpression:

		return g.generatePrintIntExpression(callExpr)

	case *ast.BooleanLiteral:

		return g.generatePrintBoolExpression(callExpr)

	case *ast.CallExpression:

		return g.generatePrintCallExpression(callExpr, argExpr)

	case *ast.Identifier:

		return g.generatePrintIdentifier(callExpr, argExpr)

	default:

		return nil, WrapPrintComplexExpr(argExpr)
	}
}

// generateInputCall handles input function calls.
func (g *LLVMGenerator) generateInputCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != 0 {
		return nil, WrapInputWrongArgs(len(callExpr.Arguments))
	}

	// Create format string for scanf("%ld", &var)
	formatStr := constant.NewCharArrayFromString(FormatStringInt)
	formatGlobal := g.module.NewGlobalDef("", formatStr)
	formatPtr := g.builder.NewGetElementPtr(formatStr.Typ, formatGlobal,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	// Allocate space for the input integer
	inputVar := g.builder.NewAlloca(types.I64)

	// Call scanf
	scanf := g.functions["scanf"]
	g.builder.NewCall(scanf, formatPtr, inputVar)

	// Load the result

	return g.builder.NewLoad(types.I64, inputVar), nil
}

// generateUserFunctionCall handles user-defined function calls.
func (g *LLVMGenerator) generateUserFunctionCall(
	funcName string,
	fn *ir.Func,
	callExpr *ast.CallExpression,
) (value.Value, error) {
	// Check if function has multiple parameters and enforce named arguments
	params, paramExists := g.functionParameters[funcName]
	if paramExists && len(params) > 1 {
		// Multi-parameter function - MUST use named arguments
		if len(callExpr.NamedArguments) == 0 {
			example := g.buildNamedArgumentsExample(params)

			return nil, WrapFunctionRequiresNamed(funcName, len(params), example)
		}

		if len(callExpr.Arguments) > 0 {
			example := g.buildNamedArgumentsExample(params)

			return nil, WrapFunctionRequiresNamed(funcName, len(params), example)
		}
	}

	// Handle named arguments vs positional arguments
	if len(callExpr.NamedArguments) > 0 {
		// Validate named arguments are not of type 'any'
		for _, namedArg := range callExpr.NamedArguments {
			if err := g.validateNotAnyType(namedArg.Value, AnyOpFunctionArgument); err != nil {
				return nil, WrapAnyDirectFunctionArg(funcName, "unknown")
			}
		}

		// Named arguments - need to reorder them to match function signature
		args, err := g.reorderNamedArguments(funcName, callExpr.NamedArguments)
		if err != nil {
			return nil, err
		}

		return g.builder.NewCall(fn, args...), nil
	}

	// Positional arguments (traditional)
	args := make([]value.Value, len(callExpr.Arguments))

	for i, arg := range callExpr.Arguments {
		// Validate that argument is not of type 'any'
		if err := g.validateNotAnyType(arg, AnyOpFunctionArgument); err != nil {
			return nil, WrapAnyDirectFunctionArg(funcName, "unknown")
		}

		val, err := g.generateExpression(arg)
		if err != nil {
			return nil, err
		}

		args[i] = val
	}

	return g.builder.NewCall(fn, args...), nil
}

// generatePrintStringLiteral handles printing string literals.
func (g *LLVMGenerator) generatePrintStringLiteral(callExpr *ast.CallExpression) (value.Value, error) {
	arg, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	puts := g.functions["puts"]
	result := g.builder.NewCall(puts, arg)

	return g.builder.NewSExt(result, types.I64), nil
}

// generatePrintIntExpression handles printing integer expressions.
func (g *LLVMGenerator) generatePrintIntExpression(callExpr *ast.CallExpression) (value.Value, error) {
	arg, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	stringArg, err := g.generateIntToString(arg)
	if err != nil {
		return nil, err
	}

	puts := g.functions["puts"]
	result := g.builder.NewCall(puts, stringArg)

	return g.builder.NewSExt(result, types.I64), nil
}

// generatePrintBoolExpression handles printing boolean expressions.
func (g *LLVMGenerator) generatePrintBoolExpression(callExpr *ast.CallExpression) (value.Value, error) {
	arg, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	stringArg, err := g.generateBoolToString(arg)
	if err != nil {
		return nil, err
	}

	puts := g.functions["puts"]
	result := g.builder.NewCall(puts, stringArg)

	return g.builder.NewSExt(result, types.I64), nil
}

// generatePrintCallExpression handles printing call expressions.
func (g *LLVMGenerator) generatePrintCallExpression(
	callExpr *ast.CallExpression,
	argExpr *ast.CallExpression,
) (value.Value, error) {
	if callIdent, ok := argExpr.Function.(*ast.Identifier); ok {
		if callIdent.Name == ToStringFunc {
			// toString call - allowed directly
			return g.generatePrintStringLiteral(callExpr)
		}

		if returnType, exists := g.functionReturnTypes[callIdent.Name]; exists {
			switch returnType {
			case TypeString:
				return g.generatePrintStringLiteral(callExpr)
			case TypeInt:
				return g.generatePrintIntExpression(callExpr)
			case TypeBool:
				return g.generatePrintBoolExpression(callExpr)
			case TypeAny:
				// 'any' type is represented as i64 at LLVM level, treat as int
				return g.generatePrintIntExpression(callExpr)
			default:
				return nil, WrapPrintCannotConvertFunc(returnType, callIdent.Name)
			}
		}

		return nil, WrapPrintUnknownFunc(callIdent.Name)
	}

	return nil, ErrPrintComplexExpr
}

// generatePrintIdentifier handles printing identifier expressions.
func (g *LLVMGenerator) generatePrintIdentifier(
	callExpr *ast.CallExpression,
	argExpr *ast.Identifier,
) (value.Value, error) {
	if varType, exists := g.variableTypes[argExpr.Name]; exists {
		switch varType {
		case TypeString:
			return g.generatePrintStringLiteral(callExpr)
		case TypeInt:
			return g.generatePrintIntExpression(callExpr)
		case TypeBool:
			return g.generatePrintBoolExpression(callExpr)
		case TypeAny:
			// 'any' type is represented as i64 at LLVM level, treat as int
			return g.generatePrintIntExpression(callExpr)
		default:
			return nil, WrapPrintCannotConvert(argExpr.Name, varType)
		}
	}

	// Unknown variable type - assume int and try auto-conversion

	return g.generatePrintIntExpression(callExpr)
}

// generateInterpolatedString generates LLVM IR for interpolated strings by concatenating parts.
func (g *LLVMGenerator) generateInterpolatedString(interpStr *ast.InterpolatedStringLiteral) (value.Value, error) {
	// For now, we'll use a simple approach: build the string by calling printf multiple times
	// into a buffer. A more sophisticated implementation would use string concatenation.
	// If there's only one part and it's text, treat it as a regular string
	if len(interpStr.Parts) == 1 && !interpStr.Parts[0].IsExpression {
		str := constant.NewCharArrayFromString(interpStr.Parts[0].Text + StringTerminator)
		global := g.module.NewGlobalDef("", str)

		return g.builder.NewGetElementPtr(str.Typ, global,
			constant.NewInt(types.I32, ArrayIndexZero),
			constant.NewInt(types.I32, ArrayIndexZero)), nil
	}

	// For multiple parts or expressions, we'll use sprintf to build the string
	// First, we need sprintf declared
	sprintf := g.ensureSprintfDeclaration()

	// Build format string and collect arguments
	var formatParts []string

	var args []value.Value

	for _, part := range interpStr.Parts {
		if part.IsExpression {
			// Auto-call toString() on all expressions in string interpolation
			toStringCall := &ast.CallExpression{
				Function:  &ast.Identifier{Name: "toString"},
				Arguments: []ast.Expression{part.Expression},
			}

			// Generate the toString call which will return a string
			val, err := g.generateExpression(toStringCall)
			if err != nil {
				return nil, err
			}

			args = append(args, val)

			// All expressions become %s since toString() always returns string
			formatParts = append(formatParts, "%s")
		} else {
			// Escape % characters in literal text by replacing % with %%
			escapedText := strings.ReplaceAll(part.Text, "%", "%%")
			formatParts = append(formatParts, escapedText)
		}
	}

	// Create the format string
	formatString := strings.Join(formatParts, "") + StringTerminator
	formatStr := constant.NewCharArrayFromString(formatString)
	formatGlobal := g.module.NewGlobalDef("", formatStr)
	formatPtr := g.builder.NewGetElementPtr(formatStr.Typ, formatGlobal,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	// Allocate buffer for result string (simplified - use fixed size)
	bufferType := types.NewArray(BufferSize1KB, types.I8) // 1KB buffer
	buffer := g.builder.NewAlloca(bufferType)
	bufferPtr := g.builder.NewGetElementPtr(bufferType, buffer,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	// Call sprintf(buffer, format, args...)
	sprintfArgs := make([]value.Value, 0, len(args)+TwoArgs)
	sprintfArgs = append(sprintfArgs, bufferPtr, formatPtr)
	sprintfArgs = append(sprintfArgs, args...)

	g.builder.NewCall(sprintf, sprintfArgs...)

	return bufferPtr, nil
}

// ensureSprintfDeclaration declares sprintf if not already declared.
func (g *LLVMGenerator) ensureSprintfDeclaration() *ir.Func {
	if sprintf, exists := g.functions["sprintf"]; exists {
		return sprintf
	}

	// Declare sprintf: i32 @sprintf(i8* %str, i8* %format, ...)
	sprintf := g.module.NewFunc("sprintf", types.I32,
		ir.NewParam("str", types.I8Ptr),
		ir.NewParam("format", types.I8Ptr))
	sprintf.Sig.Variadic = true
	g.functions["sprintf"] = sprintf

	return sprintf
}

func (g *LLVMGenerator) generateIntToString(arg value.Value) (value.Value, error) {
	// Ensure sprintf is declared
	sprintf := g.ensureSprintfDeclaration()

	// Create format string for integer conversion
	formatStr := constant.NewCharArrayFromString("%ld\x00")
	formatGlobal := g.module.NewGlobalDef("", formatStr)
	formatPtr := g.builder.NewGetElementPtr(formatStr.Typ, formatGlobal,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	// Allocate buffer for result string (64 bytes should be enough for any 64-bit integer)
	bufferType := types.NewArray(BufferSize64Bytes, types.I8)
	buffer := g.builder.NewAlloca(bufferType)
	bufferPtr := g.builder.NewGetElementPtr(bufferType, buffer,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	// Call sprintf(buffer, "%ld", arg)
	g.builder.NewCall(sprintf, bufferPtr, formatPtr, arg)

	return bufferPtr, nil
}

func (g *LLVMGenerator) generateBoolToString(arg value.Value) (value.Value, error) {
	// Create blocks for true/false cases
	blockSuffix := fmt.Sprintf("_%p", arg) // Use pointer address for uniqueness
	currentBlock := g.builder

	trueBlock := g.function.NewBlock("bool_true" + blockSuffix)
	falseBlock := g.function.NewBlock("bool_false" + blockSuffix)
	endBlock := g.function.NewBlock("bool_end" + blockSuffix)

	// Check if arg == 1 (true) or 0 (false)
	zero := constant.NewInt(types.I64, ArrayIndexZero)
	isTrue := currentBlock.NewICmp(enum.IPredNE, arg, zero)
	currentBlock.NewCondBr(isTrue, trueBlock, falseBlock)

	// True case - return "true"
	g.builder = trueBlock
	trueStr := constant.NewCharArrayFromString(TrueString)
	trueGlobal := g.module.NewGlobalDef("", trueStr)
	truePtr := trueBlock.NewGetElementPtr(trueStr.Typ, trueGlobal,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	trueBlock.NewBr(endBlock)

	// False case - return "false"
	g.builder = falseBlock
	falseStr := constant.NewCharArrayFromString(FalseString)
	falseGlobal := g.module.NewGlobalDef("", falseStr)
	falsePtr := falseBlock.NewGetElementPtr(falseStr.Typ, falseGlobal,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	falseBlock.NewBr(endBlock)

	// Create PHI node in end block
	g.builder = endBlock
	phi := endBlock.NewPhi(ir.NewIncoming(truePtr, trueBlock), ir.NewIncoming(falsePtr, falseBlock))

	return phi, nil
}

func (g *LLVMGenerator) generateMatchExpression(matchExpr *ast.MatchExpression) (value.Value, error) {
	// Validate match expression for exhaustiveness and unknown variants
	if err := g.validateMatchExpression(matchExpr); err != nil {
		return nil, err
	}

	discriminant, err := g.generateExpression(matchExpr.Expression)
	if err != nil {
		return nil, err
	}

	return g.generateMatchExpressionWithDiscriminant(matchExpr, discriminant)
}

// generateMatchExpressionWithDiscriminant generates match expression with pre-computed discriminant.
func (g *LLVMGenerator) generateMatchExpressionWithDiscriminant(
	matchExpr *ast.MatchExpression,
	discriminant value.Value,
) (value.Value, error) {
	if g.hasResultPatterns(matchExpr.Arms) {
		return g.generateResultMatchExpression(matchExpr, discriminant)
	}

	if len(matchExpr.Arms) == 0 {
		return constant.NewInt(types.I64, ArrayIndexZero), nil
	}

	return g.generateStandardMatchExpression(matchExpr, discriminant)
}

// hasResultPatterns checks if the match expression has Success/Err patterns.
func (g *LLVMGenerator) hasResultPatterns(arms []ast.MatchArm) bool {
	for _, arm := range arms {
		if arm.Pattern.Constructor == "Success" || arm.Pattern.Constructor == "Err" {
			return true
		}
	}

	return false
}

// generateStandardMatchExpression generates a standard (non-result) match expression.
func (g *LLVMGenerator) generateStandardMatchExpression(
	matchExpr *ast.MatchExpression,
	discriminant value.Value,
) (value.Value, error) {
	// Create unique block names for this match expression
	blockSuffix := fmt.Sprintf("_%p", matchExpr)
	endBlock := g.function.NewBlock("match_end" + blockSuffix)
	armBlocks := g.createMatchArmBlocks(matchExpr.Arms, blockSuffix)

	// Generate conditions first (creates branching logic)
	g.generateMatchConditions(matchExpr.Arms, armBlocks, discriminant, blockSuffix)

	// Then generate arm values (adds terminators to arm blocks)
	armValues, predecessorBlocks, err := g.generateMatchArmValues(matchExpr.Arms, armBlocks, endBlock, discriminant)
	if err != nil {
		return nil, err
	}

	return g.createMatchResult(armValues, predecessorBlocks, endBlock)
}

// createMatchArmBlocks creates LLVM blocks for each match arm.
func (g *LLVMGenerator) createMatchArmBlocks(arms []ast.MatchArm, blockSuffix string) []*ir.Block {
	var armBlocks []*ir.Block
	for i := range arms {
		armBlocks = append(armBlocks, g.function.NewBlock(fmt.Sprintf("match_arm_%d%s", i, blockSuffix)))
	}

	return armBlocks
}

// generateMatchArmValues generates values for each match arm.
func (g *LLVMGenerator) generateMatchArmValues(
	arms []ast.MatchArm,
	armBlocks []*ir.Block,
	endBlock *ir.Block,
	discriminant value.Value,
) ([]value.Value, []*ir.Block, error) {
	var armValues []value.Value
	var predecessorBlocks []*ir.Block
	oldBuilder := g.builder

	for i, arm := range arms {
		g.builder = armBlocks[i]

		// Handle variable binding in patterns
		if arm.Pattern.Variable != "" {
			// Save the current variable scope
			oldVariables := make(map[string]value.Value)
			for k, v := range g.variables {
				oldVariables[k] = v
			}

			// Bind the pattern variable to the discriminant value
			g.variables[arm.Pattern.Variable] = discriminant

			// Generate the arm expression
			armValue, err := g.generateExpression(arm.Expression)
			if err != nil {
				return nil, nil, err
			}

			// Restore the previous variable scope
			g.variables = oldVariables

			armValues = append(armValues, armValue)
		} else {
			// No variable binding, generate normally
			armValue, err := g.generateExpression(arm.Expression)
			if err != nil {
				return nil, nil, err
			}

			armValues = append(armValues, armValue)
		}

		armBlocks[i].NewBr(endBlock)
		predecessorBlocks = append(predecessorBlocks, armBlocks[i])
	}

	g.builder = oldBuilder

	return armValues, predecessorBlocks, nil
}

// generateMatchConditions generates the conditional branches for pattern matching.
func (g *LLVMGenerator) generateMatchConditions(
	arms []ast.MatchArm,
	armBlocks []*ir.Block,
	discriminant value.Value,
	blockSuffix string,
) {
	currentBlock := g.builder

	for i, arm := range arms {
		// Ensure the builder is set to the current block
		g.builder = currentBlock
		condition := g.createPatternCondition(arm.Pattern, discriminant, currentBlock)

		if i == len(arms)-1 {
			currentBlock.NewBr(armBlocks[i])
		} else {
			nextCheckBlock := g.function.NewBlock(fmt.Sprintf("match_check_%d%s", i+1, blockSuffix))
			currentBlock.NewCondBr(condition, armBlocks[i], nextCheckBlock)
			currentBlock = nextCheckBlock
		}
	}
}

// createPatternCondition creates a condition for pattern matching.
func (g *LLVMGenerator) createPatternCondition(
	pattern ast.Pattern,
	discriminant value.Value,
	currentBlock *ir.Block,
) value.Value {
	if pattern.Constructor == "_" || pattern.Constructor == UnknownPattern {
		return constant.NewBool(true)
	}

	// Handle variable binding patterns (empty constructor means variable binding)
	if pattern.Constructor == "" && pattern.Variable != "" {
		return constant.NewBool(true)
	}

	// Check if it's a union type variant
	if discriminantValue, exists := g.unionVariants[pattern.Constructor]; exists {
		patternConst := constant.NewInt(types.I64, discriminantValue)

		return currentBlock.NewICmp(enum.IPredEQ, discriminant, patternConst)
	}

	// Check if it's a numeric literal
	if patternValue, err := strconv.ParseInt(pattern.Constructor, 10, 64); err == nil {
		patternConst := constant.NewInt(types.I64, patternValue)

		return currentBlock.NewICmp(enum.IPredEQ, discriminant, patternConst)
	}

	return g.createStringPatternCondition(pattern.Constructor, discriminant, currentBlock)
}

// createStringPatternCondition creates a condition for string pattern matching.
func (g *LLVMGenerator) createStringPatternCondition(
	constructor string,
	discriminant value.Value,
	currentBlock *ir.Block,
) value.Value {
	if !strings.HasPrefix(constructor, "\"") || !strings.HasSuffix(constructor, "\"") {
		return constant.NewBool(false)
	}

	patternStr := constructor[1 : len(constructor)-1]
	patternConstStr := constant.NewCharArrayFromString(patternStr + StringTerminator)
	patternGlobal := g.module.NewGlobalDef("", patternConstStr)
	patternPtr := currentBlock.NewGetElementPtr(patternConstStr.Typ, patternGlobal,
		constant.NewInt(types.I32, ArrayIndexZero), constant.NewInt(types.I32, ArrayIndexZero))

	strcmp := g.functions["strcmp"]
	compareResult := currentBlock.NewCall(strcmp, discriminant, patternPtr)
	zero := constant.NewInt(types.I32, ArrayIndexZero)

	return currentBlock.NewICmp(enum.IPredEQ, compareResult, zero)
}

// createMatchResult creates the final result value using PHI nodes if needed.
func (g *LLVMGenerator) createMatchResult(
	armValues []value.Value,
	predecessorBlocks []*ir.Block,
	endBlock *ir.Block,
) (value.Value, error) {
	g.builder = endBlock

	if len(armValues) == 1 {
		return armValues[0], nil
	}

	// Check if we need type coercion
	coercedValues, err := g.coerceArmValuesToCommonType(armValues)
	if err != nil {
		return nil, err
	}

	var incomings []*ir.Incoming
	for i, val := range coercedValues {
		incomings = append(incomings, ir.NewIncoming(val, predecessorBlocks[i]))
	}

	return endBlock.NewPhi(incomings...), nil
}

// coerceArmValuesToCommonType ensures all arm values have compatible types.
func (g *LLVMGenerator) coerceArmValuesToCommonType(armValues []value.Value) ([]value.Value, error) {
	expectedType := armValues[0].Type()

	if !g.needsTypeCoercion(armValues, expectedType) {
		return armValues, nil
	}

	return g.performTypeCoercion(armValues, expectedType)
}

// needsTypeCoercion checks if type coercion is needed for arm values.
func (g *LLVMGenerator) needsTypeCoercion(armValues []value.Value, expectedType types.Type) bool {
	for i := 1; i < len(armValues); i++ {
		if armValues[i].Type() != expectedType {
			return true
		}
	}

	return false
}

// performTypeCoercion converts all values to the expected type.
func (g *LLVMGenerator) performTypeCoercion(armValues []value.Value, expectedType types.Type) ([]value.Value, error) {
	coercedValues := make([]value.Value, len(armValues))

	for i, val := range armValues {
		if val.Type() == expectedType {
			coercedValues[i] = val
		} else {
			coercedVal, err := g.coerceValueToType(val, expectedType)
			if err != nil {
				return nil, err
			}
			coercedValues[i] = coercedVal
		}
	}

	return coercedValues, nil
}

// coerceValueToType converts a single value to the target type.
func (g *LLVMGenerator) coerceValueToType(val value.Value, targetType types.Type) (value.Value, error) {
	switch targetType {
	case types.I8Ptr:

		return g.convertToString(val)
	case types.I64:

		return constant.NewInt(types.I64, 0), nil
	default:

		return val, nil
	}
}

// convertToString converts a value to string type.
func (g *LLVMGenerator) convertToString(val value.Value) (value.Value, error) {
	if val.Type() == types.I64 {
		return g.generateIntToString(val)
	}

	return val, nil
}

func (g *LLVMGenerator) generateResultMatchExpression(
	matchExpr *ast.MatchExpression,
	discriminant value.Value,
) (value.Value, error) {
	blocks := g.createResultMatchBlocks(matchExpr)
	g.generateResultMatchCondition(discriminant, blocks)

	successValue, err := g.generateSuccessBlock(matchExpr, blocks)
	if err != nil {
		return nil, err
	}

	errorValue, err := g.generateErrorBlock(matchExpr, blocks)
	if err != nil {
		return nil, err
	}

	return g.createResultMatchPhi(successValue, errorValue, blocks)
}

// ResultMatchBlocks holds the blocks for result match expressions.
type ResultMatchBlocks struct {
	Success *ir.Block
	Error   *ir.Block
	End     *ir.Block
}

// createResultMatchBlocks creates blocks for result match expressions.
func (g *LLVMGenerator) createResultMatchBlocks(matchExpr *ast.MatchExpression) *ResultMatchBlocks {
	blockSuffix := fmt.Sprintf("_%p", matchExpr)

	return &ResultMatchBlocks{
		Success: g.function.NewBlock("success_case" + blockSuffix),
		Error:   g.function.NewBlock("error_case" + blockSuffix),
		End:     g.function.NewBlock("match_end" + blockSuffix),
	}
}

// generateResultMatchCondition generates the condition for result matching.
func (g *LLVMGenerator) generateResultMatchCondition(discriminant value.Value, blocks *ResultMatchBlocks) {
	zero := constant.NewInt(types.I64, ArrayIndexZero)
	isSuccess := g.builder.NewICmp(enum.IPredSGE, discriminant, zero)
	g.builder.NewCondBr(isSuccess, blocks.Success, blocks.Error)
}

// generateSuccessBlock generates the success block for result matching.
func (g *LLVMGenerator) generateSuccessBlock(
	matchExpr *ast.MatchExpression,
	blocks *ResultMatchBlocks,
) (value.Value, error) {
	g.builder = blocks.Success

	successExpr := g.findSuccessValue(matchExpr)
	var successValue value.Value
	if successExpr == nil {
		successValue = constant.NewInt(types.I64, ArrayIndexZero)
	} else {
		val, err := g.generateExpression(successExpr)
		if err != nil {
			return nil, err
		}
		successValue = val
	}

	blocks.Success.NewBr(blocks.End)

	return successValue, nil
}

// generateErrorBlock generates the error block for result matching.
func (g *LLVMGenerator) generateErrorBlock(
	matchExpr *ast.MatchExpression,
	blocks *ResultMatchBlocks,
) (value.Value, error) {
	g.builder = blocks.Error

	errorExpr := g.findErrorValue(matchExpr)
	var errorValue value.Value
	if errorExpr == nil {
		errorValue = constant.NewInt(types.I64, ArrayIndexZero)
	} else {
		val, err := g.generateExpression(errorExpr)
		if err != nil {
			return nil, err
		}
		errorValue = val
	}

	blocks.Error.NewBr(blocks.End)

	return errorValue, nil
}

// findSuccessValue finds the success expression in match arms.
func (g *LLVMGenerator) findSuccessValue(matchExpr *ast.MatchExpression) ast.Expression {
	for _, arm := range matchExpr.Arms {
		if arm.Pattern.Constructor == "Success" {
			return arm.Expression
		}
	}

	return nil
}

// findErrorValue finds the error expression in match arms.
func (g *LLVMGenerator) findErrorValue(matchExpr *ast.MatchExpression) ast.Expression {
	for _, arm := range matchExpr.Arms {
		if arm.Pattern.Constructor == "Err" {
			return arm.Expression
		}
	}

	return nil
}

// createResultMatchPhi creates the PHI node for result matching.
func (g *LLVMGenerator) createResultMatchPhi(
	successValue, errorValue value.Value,
	blocks *ResultMatchBlocks,
) (value.Value, error) {
	g.builder = blocks.End

	// Ensure both values have the same type for PHI node
	if successValue.Type() != errorValue.Type() {
		// Determine which type to use - prefer string if either is string
		if successValue.Type() == types.I8Ptr {
			// Convert errorValue to string if it's not already
			if errorValue.Type() == types.I64 {
				convertedError, err := g.generateIntToString(errorValue)
				if err != nil {
					return nil, err
				}
				errorValue = convertedError
			}
		} else if errorValue.Type() == types.I8Ptr {
			// Convert successValue to string if it's not already
			if successValue.Type() == types.I64 {
				convertedSuccess, err := g.generateIntToString(successValue)
				if err != nil {
					return nil, err
				}
				successValue = convertedSuccess
			}
		}
		// If both are different non-string types, convert both to int
		if successValue.Type() != errorValue.Type() {
			successValue = constant.NewInt(types.I64, 0)
			errorValue = constant.NewInt(types.I64, 0)
		}
	}

	phi := blocks.End.NewPhi(
		ir.NewIncoming(successValue, blocks.Success),
		ir.NewIncoming(errorValue, blocks.Error),
	)

	return phi, nil
}

// generateRangeCall handles range function calls - creates an iterator from start to end.
func (g *LLVMGenerator) generateRangeCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != TwoArgs {
		return nil, WrapRangeWrongArgs(len(callExpr.Arguments))
	}

	start, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	end, err := g.generateExpression(callExpr.Arguments[1])
	if err != nil {
		return nil, err
	}

	// Create a struct to hold range data: {start, end}
	rangeStructType := types.NewStruct(types.I64, types.I64)
	rangeValue := g.builder.NewAlloca(rangeStructType)

	// Store start value at index 0
	startPtr := g.builder.NewGetElementPtr(rangeStructType, rangeValue,
		constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0))
	g.builder.NewStore(start, startPtr)

	// Store end value at index 1
	endPtr := g.builder.NewGetElementPtr(rangeStructType, rangeValue,
		constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 1))
	g.builder.NewStore(end, endPtr)

	return rangeValue, nil
}

// generateForEachCall handles forEach function calls - applies a function to each element.
func (g *LLVMGenerator) generateForEachCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != TwoArgs {
		return nil, WrapForEachWrongArgs(len(callExpr.Arguments))
	}

	// Get the range struct from first argument
	rangeValue, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	// Get the function to apply from second argument
	funcArg := callExpr.Arguments[1]
	funcIdent, ok := funcArg.(*ast.Identifier)
	if !ok {
		return nil, ErrForEachNotFunction
	}

	// Extract range bounds
	start, end := g.extractRangeBounds(rangeValue)

	// Create loop blocks
	blocks := g.createForEachLoopBlocks(callExpr)

	// Generate loop logic
	err = g.generateForEachLoop(start, end, funcIdent, blocks)
	if err != nil {
		return nil, err
	}

	// Return the original range struct for potential pipe chaining

	return rangeValue, nil
}

// extractRangeBounds extracts start and end values from a range struct.
func (g *LLVMGenerator) extractRangeBounds(rangeValue value.Value) (value.Value, value.Value) {
	// Define the range struct type
	rangeStructType := types.NewStruct(types.I64, types.I64)

	// Check if rangeValue is already a pointer to the struct or if we need to handle it differently
	var startPtr, endPtr value.Value

	// Check the type of rangeValue - if it's a pointer to struct, use it directly
	// If it's the struct value itself, we need to handle it differently
	if rangeValue.Type().String() == rangeStructType.String()+"*" {
		// It's a pointer to the struct (from range() call)
		startPtr = g.builder.NewGetElementPtr(rangeStructType, rangeValue,
			constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0))
		endPtr = g.builder.NewGetElementPtr(rangeStructType, rangeValue,
			constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 1))
	} else {
		// It's likely the struct value itself or something else - allocate and store
		tempRange := g.builder.NewAlloca(rangeStructType)
		g.builder.NewStore(rangeValue, tempRange)

		startPtr = g.builder.NewGetElementPtr(rangeStructType, tempRange,
			constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0))
		endPtr = g.builder.NewGetElementPtr(rangeStructType, tempRange,
			constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 1))
	}

	start := g.builder.NewLoad(types.I64, startPtr)
	end := g.builder.NewLoad(types.I64, endPtr)

	return start, end
}

// ForEachLoopBlocks holds the basic blocks for a forEach loop.
type ForEachLoopBlocks struct {
	LoopCond *ir.Block
	LoopBody *ir.Block
	LoopEnd  *ir.Block
}

// createForEachLoopBlocks creates the basic blocks needed for a forEach loop.
func (g *LLVMGenerator) createForEachLoopBlocks(callExpr *ast.CallExpression) *ForEachLoopBlocks {
	// Create unique block names using pointer address
	blockSuffix := fmt.Sprintf("_%p", callExpr)

	return &ForEachLoopBlocks{
		LoopCond: g.function.NewBlock("loop_cond" + blockSuffix),
		LoopBody: g.function.NewBlock("loop_body" + blockSuffix),
		LoopEnd:  g.function.NewBlock("loop_end" + blockSuffix),
	}
}

// generateForEachLoop generates the actual loop logic for forEach.
func (g *LLVMGenerator) generateForEachLoop(
	start, end value.Value,
	funcIdent *ast.Identifier,
	blocks *ForEachLoopBlocks,
) error {
	// Allocate and initialize the loop counter
	counterPtr := g.builder.NewAlloca(types.I64)
	g.builder.NewStore(start, counterPtr)

	// Branch to loop condition check
	g.builder.NewBr(blocks.LoopCond)

	// Loop condition: while (counter < end)
	g.builder = blocks.LoopCond
	currentCounter := g.builder.NewLoad(types.I64, counterPtr)
	condition := g.builder.NewICmp(enum.IPredSLT, currentCounter, end)
	g.builder.NewCondBr(condition, blocks.LoopBody, blocks.LoopEnd)

	// Loop body: call function with current counter value
	g.builder = blocks.LoopBody
	counterValue := g.builder.NewLoad(types.I64, counterPtr)

	// Call the function with the current counter value
	_, err := g.callFunctionWithValue(funcIdent, counterValue)
	if err != nil {
		return err
	}

	// Increment the counter: counter = counter + 1
	one := constant.NewInt(types.I64, 1)
	incrementedValue := g.builder.NewAdd(counterValue, one)
	g.builder.NewStore(incrementedValue, counterPtr)

	// Branch back to condition check
	g.builder.NewBr(blocks.LoopCond)

	// After the loop
	g.builder = blocks.LoopEnd

	return nil
}

// generateMapCall handles map function calls - transforms each element.
func (g *LLVMGenerator) generateMapCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != TwoArgs {
		return nil, WrapMapWrongArgs(len(callExpr.Arguments))
	}

	// Get the range iterator
	rangeValue, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	// Get the function to apply
	funcArg := callExpr.Arguments[1]
	if _, ok := funcArg.(*ast.Identifier); !ok {
		return nil, ErrMapNotFunction
	}

	// For now, map just returns the original range since we're dealing with lazy evaluation
	// In a full implementation, this would create a new iterator that applies the function
	// when iterated over. For simplicity, we'll return the range and let forEach handle
	// the function application.

	// TODO: Implement proper lazy map that stores the transformation function
	// For now, just return the range and the transformation will happen in forEach

	return rangeValue, nil
}

// generateFilterCall handles filter function calls - selects elements based on predicate.
func (g *LLVMGenerator) generateFilterCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != TwoArgs {
		return nil, WrapFilterWrongArgs(len(callExpr.Arguments))
	}

	iterator, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	// Get the predicate function
	funcArg := callExpr.Arguments[1]
	if funcIdent, ok := funcArg.(*ast.Identifier); ok {
		// Apply the predicate and return the result
		return g.callFunctionWithValue(funcIdent, iterator)
	}

	return nil, ErrFilterNotFunction
}

// generateFoldCall handles fold function calls - reduces iterator to single value.
func (g *LLVMGenerator) generateFoldCall(callExpr *ast.CallExpression) (value.Value, error) {
	if len(callExpr.Arguments) != ThreeArgs {
		return nil, WrapFoldWrongArgs(len(callExpr.Arguments))
	}

	iterator, err := g.generateExpression(callExpr.Arguments[0])
	if err != nil {
		return nil, err
	}

	initial, err := g.generateExpression(callExpr.Arguments[1])
	if err != nil {
		return nil, err
	}

	// Get the fold function
	funcArg := callExpr.Arguments[2]
	if funcIdent, ok := funcArg.(*ast.Identifier); ok {
		// Apply the function with both accumulator and current value
		return g.callFunctionWithTwoValues(funcIdent, initial, iterator)
	}

	return nil, ErrFoldNotFunction
}

// callFunctionWithValue calls any function (built-in or user-defined) with a single value argument.
func (g *LLVMGenerator) callFunctionWithValue(
	funcIdent *ast.Identifier,
	value value.Value,
) (value.Value, error) {
	// Handle built-in functions
	switch funcIdent.Name {
	case PrintFunc:

		return g.callBuiltInPrint(value)
	case ToStringFunc:

		return g.callBuiltInToString(value)
	case InputFunc:

		return nil, ErrInputNoArgs
	case "testAny":
		// Test function that returns 'any' type to test validation
		return g.generateTestAnyCall()
	}

	// Handle user-defined functions
	if fn, exists := g.functions[funcIdent.Name]; exists {
		return g.builder.NewCall(fn, value), nil
	}

	return nil, WrapFunctionNotFound(funcIdent.Name)
}

// callFunctionWithTwoValues calls any function with two value arguments.
func (g *LLVMGenerator) callFunctionWithTwoValues(
	funcIdent *ast.Identifier,
	value1, value2 value.Value,
) (value.Value, error) {
	// Built-in functions typically don't take two arguments, but handle edge cases
	switch funcIdent.Name {
	case PrintFunc, ToStringFunc, InputFunc:

		return nil, WrapBuiltInTwoArgs(funcIdent.Name)
	}

	// Handle user-defined functions
	if fn, exists := g.functions[funcIdent.Name]; exists {
		return g.builder.NewCall(fn, value1, value2), nil
	}

	return nil, WrapFunctionNotFound(funcIdent.Name)
}

// callBuiltInPrint handles calling the built-in print function with a value.
func (g *LLVMGenerator) callBuiltInPrint(value value.Value) (value.Value, error) {
	// Convert the value to string and call puts
	stringArg, err := g.generateIntToString(value)
	if err != nil {
		return nil, err
	}

	puts := g.functions["puts"]
	result := g.builder.NewCall(puts, stringArg)

	return g.builder.NewSExt(result, types.I64), nil
}

// callBuiltInToString handles calling the built-in toString function with a value.
func (g *LLVMGenerator) callBuiltInToString(value value.Value) (value.Value, error) {
	// For now, assume integer values and convert to string

	return g.generateIntToString(value)
}

// generateTestAnyCall generates a function that returns 'any' type for testing validation.
func (g *LLVMGenerator) generateTestAnyCall() (value.Value, error) {
	// Return a simple integer but mark the function as returning 'any'
	// This will be used to test that direct access to 'any' types fails
	g.functionReturnTypes["testAny"] = TypeAny

	return constant.NewInt(types.I64, DefaultPlaceholder), nil
}
