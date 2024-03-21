package async

import (
	"context"
	"github.com/kozhurkin/async/pipers"
)

// tests: ✅
// bench: ⚠️ too slow

func AsyncPipers[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	pp := make(pipers.Pipers[V], len(args))

	for i, a := range args {
		i, a := i, a
		pp[i] = pipers.NewPiper(func() (V, error) {
			printDebug("++ call i=%v, a=%v", i, a)
			value, err := f(i, a)
			printDebug("^ defer i=%v, a=%v, err=%v", i, a, err)
			return value, err
		})
	}

	pp.RunConcurrency(concurrency)

	select {
	case err := <-pp.ErrorsChan():
		return pp.Results(), err
	case <-ctx.Done():
		return pp.Results(), ctx.Err()
	}
}
