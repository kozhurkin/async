package async

import (
	"context"
)

// throws "context canceled" if an error occurs before/after cancellation: NO/NO
// does not wait for parallel jobs when an error occurs or canceled: NO
func AsyncPromise22[A any, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	promises := make([]chan V, len(args))

	traffic := make(chan struct{}, concurrency-1)

LOOP:
	for i, arg := range args {
		i, arg := i, arg
		promises[i] = Pipeline(func() V {
			printDebug("JOB START: i=%v arg=%v", i, arg)
			value, err := f(arg)
			if err != nil {
				cancel()
			}
			printDebug("JOB DONE: i=%v arg=%v value=%v err=%v", i, arg, value, err)
			<-traffic
			return value
		})
		printDebug("promises[%v] = p", i)
		select {
		case <-ctx.Done():
			promises = promises[0:i]
			printDebug("SKIP %v", len(promises))
			break LOOP
		case traffic <- struct{}{}:
		}
	}
	printDebug("LOOP END")
	printDebug("close(traffic) %v", len(traffic))
	close(traffic)

	res := make([]V, len(args))
	for i, p := range promises {
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case msg := <-p:
			printDebug("fill %v %v %v", i, p, msg)
			res[i] = msg
		}
	}

	return res, nil
}
