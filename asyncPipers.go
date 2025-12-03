package async

import (
	"context"
	"github.com/kozhurkin/pipers"
)

// tests: ✅
// bench: ⚠️ too slow
func AsyncPipers[A any, V any](ctx context.Context, args []A, fn func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	return pipers.FromArgsCtx(args, func(ctx context.Context, i int, a A) (V, error) {
		return fn(ctx, i, a)
	}).Context(ctx).Concurrency(concurrency).Resolve()
}
