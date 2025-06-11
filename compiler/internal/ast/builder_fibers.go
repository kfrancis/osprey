package ast

import (
	"github.com/christianfindlay/osprey/parser"
)

// buildSpawnExpression builds a SpawnExpression from a spawn context.
func (b *Builder) buildSpawnExpression(ctx parser.IPrimaryContext) *SpawnExpression {
	// The spawn expression should have a child expression
	expr := b.buildExpression(ctx.Expr(0))

	return &SpawnExpression{
		Expression: expr,
	}
}

// buildAwaitExpression builds an AwaitExpression from await(fiber) syntax.
func (b *Builder) buildAwaitExpression(ctx parser.IPrimaryContext) *AwaitExpression {
	// AWAIT LPAREN expr RPAREN
	fiber := b.buildExpression(ctx.Expr(0))

	return &AwaitExpression{
		Expression: fiber,
	}
}

// buildYieldExpression builds a YieldExpression from yield syntax.
func (b *Builder) buildYieldExpression(ctx parser.IPrimaryContext) *YieldExpression {
	// YIELD expr? - optional expression
	if len(ctx.AllExpr()) > 0 {
		// yield expr
		value := b.buildExpression(ctx.Expr(0))

		return &YieldExpression{
			Value: value,
		}
	}

	// yield with no value
	return &YieldExpression{
		Value: nil,
	}
}

// buildSelectExpression builds a SelectExpression from a select context.
func (b *Builder) buildSelectExpression(ctx *parser.SelectExprContext) *SelectExpression {
	var arms []SelectArm

	for _, armCtx := range ctx.AllSelectArm() {
		pattern := b.buildPattern(armCtx.Pattern())
		expr := b.buildExpression(armCtx.Expr())

		arms = append(arms, SelectArm{
			Pattern:    pattern,
			Operation:  nil, // Operation is encoded in the pattern for now
			Expression: expr,
		})
	}

	return &SelectExpression{
		Arms: arms,
	}
}

// buildSendExpression builds a ChannelSendExpression from send(channel, value) syntax.
func (b *Builder) buildSendExpression(ctx parser.IPrimaryContext) *ChannelSendExpression {
	// SEND LPAREN expr COMMA expr RPAREN
	// The first expr is the channel, the second expr is the value
	channel := b.buildExpression(ctx.Expr(0))
	value := b.buildExpression(ctx.Expr(1))

	return &ChannelSendExpression{
		Channel: channel,
		Value:   value,
	}
}

// buildRecvExpression builds a ChannelRecvExpression from recv(channel) syntax.
func (b *Builder) buildRecvExpression(ctx parser.IPrimaryContext) *ChannelRecvExpression {
	// RECV LPAREN expr RPAREN
	channel := b.buildExpression(ctx.Expr(0))

	return &ChannelRecvExpression{
		Channel: channel,
	}
}
