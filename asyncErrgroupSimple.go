package async

import (
	"context"
	"golang.org/x/sync/errgroup"
)

func AsyncErrgroupSimple[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(concurrency)

	res := make([]V, len(args))

	for i, arg := range args {
		i, arg := i, arg // захват переменных
		eg.Go(func() error {
			value, err := f(ctx, i, arg)
			if err != nil {
				return err
			}
			res[i] = value
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return res, err
	}
	return res, nil
}
