package ast

import (
	"github.com/christianfindlay/osprey/parser"
)

// buildArguments parses argument lists, handling both positional and named arguments.
func (b *Builder) buildArguments(ctx parser.IArgListContext) ([]Expression, []NamedArgument) {
	if ctx == nil {
		return nil, nil
	}

	// Check if we have named arguments (namedArgList)
	if ctx.NamedArgList() != nil {
		return b.buildNamedArguments(ctx.NamedArgList())
	}

	// Positional arguments (traditional syntax: expr, expr...)
	if len(ctx.AllExpr()) > 0 {
		var args []Expression

		for _, exprCtx := range ctx.AllExpr() {
			expr := b.buildExpression(exprCtx)
			if expr != nil {
				args = append(args, expr)
			}
		}

		return args, nil
	}

	return nil, nil
}

// buildNamedArguments builds a list of named arguments.
func (b *Builder) buildNamedArguments(ctx parser.INamedArgListContext) ([]Expression, []NamedArgument) {
	if ctx == nil {
		return nil, nil
	}

	var namedArgs []NamedArgument

	for _, namedArgCtx := range ctx.AllNamedArg() {
		name := namedArgCtx.ID().GetText()
		value := b.buildExpression(namedArgCtx.Expr())

		namedArgs = append(namedArgs, NamedArgument{
			Name:  name,
			Value: value,
		})
	}

	return nil, namedArgs // No positional args when using named args
}
