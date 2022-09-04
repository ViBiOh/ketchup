package ketchup

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
)

type key int

const (
	ctxSuggestKey key = iota
	ctxCountKey
)

func ignoresIdsToCtx(ctx context.Context, ignoresIds []model.Identifier) context.Context {
	return context.WithValue(ctx, ctxSuggestKey, ignoresIds)
}

func ignoresIdsFromCtx(ctx context.Context) []model.Identifier {
	content := ctx.Value(ctxSuggestKey)
	values, _ := content.([]model.Identifier)

	return values
}

func countToCtx(ctx context.Context, count uint64) context.Context {
	return context.WithValue(ctx, ctxCountKey, count)
}

func countFromCtx(ctx context.Context) uint64 {
	content := ctx.Value(ctxCountKey)
	values, _ := content.(uint64)

	return values
}
