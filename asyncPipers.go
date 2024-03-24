package async

import (
	"context"
	"github.com/kozhurkin/pipers"
)

// tests: ✅
// bench: ⚠️ too slow

func AsyncPipers[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	return pipers.FromArgs(args, f).Context(ctx).Concurrency(concurrency).Resolve()
}
