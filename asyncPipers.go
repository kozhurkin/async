package async

import (
	"context"
	"github.com/kozhurkin/async/pipers"
)

// tests: ✅
// bench: ⚠️ too slow

func AsyncPipers[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	ps := pipers.PiperSolver[V]{}

	for i, a := range args {
		i, a := i, a
		ps.AddFunc(func() (V, error) {
			return f(i, a)
		})
	}

	ps.Context(ctx).Concurrency(concurrency).Run()

	return ps.Resolve()
}
