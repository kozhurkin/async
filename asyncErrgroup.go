package async

import (
	"context"
	"golang.org/x/sync/errgroup"
)

func AsyncErrgroup[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(concurrency)

	res := make([]V, len(args))

	for i, arg := range args {
		i, arg := i, arg // захват переменных
		eg.Go(func() error {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			printDebug("Task %v started", i)

			resultCh := make(chan struct {
				value V
				err   error
			}, 1)

			// Запускаем f в отдельной горутине
			go func() {
				v, err := f(ctx, i, arg)
				resultCh <- struct {
					value V
					err   error
				}{v, err}
			}()

			select {
			case <-ctx.Done():
				printDebug("Task %v canceled due to context: %v", i, ctx.Err())
				return ctx.Err()
			case r := <-resultCh:
				if r.err != nil {
					printDebug("Task %v finished with error: %v", i, r.err)
					return r.err
				}
				res[i] = r.value
				printDebug("Task %v finished successfully", i)
				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return res, err
	}
	return res, nil
}
