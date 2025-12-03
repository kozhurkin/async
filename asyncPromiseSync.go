package async

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// throws "context canceled" if an error occurs before/after cancellation: NO/YES
// instant cancellation (does not wait for parallel jobs when an error occurs or canceled): NO

func AsyncPromiseSync[A any, V any](ctx context.Context, args []A, fn func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	promises := make([]chan V, len(args))

	traffic := make(chan struct{}, concurrency-1)
	wg := new(errgroup.Group)

LOOP:
	for i, arg := range args {
		i, arg := i, arg
		select {
		case <-ctx.Done():
			promises = promises[0:i]
			printDebug("SKIP %v", len(promises))
			break LOOP
		default:
		}
		ch := make(chan V, 1)
		wg.Go(func() error {
			printDebug("JOB START: i=%v arg=%v", i, arg)
			value, err := fn(ctx, i, arg)
			printDebug("JOB DONE: i=%v arg=%v value=%v err=%v", i, arg, value, err)
			<-traffic
			ch <- value
			if err == nil {
				return nil
			}
			select {
			case <-ctx.Done():
				return nil
			default:
				cancel()
				return err
			}
		})
		promises[i] = ch
		printDebug("promises[%v] = p", i)
		traffic <- struct{}{}
	}
	printDebug("LOOP END")
	printDebug("close(traffic) %v", len(traffic))
	close(traffic)

	res := make([]V, len(args))
	for i, p := range promises {
		printDebug("fill %v %v", i, p)
		res[i] = <-p
	}

	if err := wg.Wait(); err != nil {
		return res, err
	}

	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		return res, nil
	}
}
