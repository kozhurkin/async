package async

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// tests: ✅
// bench: ✅
func AsyncErrgroup[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(concurrency)

	res := make([]V, len(args))

	for i, arg := range args {
		i, arg := i, arg
		eg.Go(func() error {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}

			resultCh := make(chan struct {
				value V
				err   error
			}, 1)

			go func() {
				v, err := f(ctx, i, arg)
				resultCh <- struct {
					value V
					err   error
				}{v, err}
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case r := <-resultCh:
				if r.err != nil {
					return r.err
				}
				res[i] = r.value
				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return res, err
	}
	return res, nil
}
