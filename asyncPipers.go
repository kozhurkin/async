package async

import (
	"context"
	"github.com/kozhurkin/pipers"
)

// tests: ✅
// bench: ⚠️ too slow

func AsyncPipers[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	return pipers.FromArgs(args, func(i int, a A) (V, error) {
		return f(ctx, i, a)
	}).Context(ctx).Concurrency(concurrency).Resolve()
}
