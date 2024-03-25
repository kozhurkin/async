package async

import (
	"context"
	"github.com/kozhurkin/async/pip"
	"sync"
)

// save the resulting array after canceling/error: NO/NO
// throws "context canceled" if an error occurs before/after cancellation: NO/NO
// instant cancellation (does not wait for parallel jobs when an error occurs or canceled): NO
func AsyncPromiseCatch[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	promises := make([]chan V, len(args))

	traffic := make(chan struct{}, concurrency-1)

	catch := make(chan error)

	var once sync.Once
	var err error

	for i, arg := range args {
		i, arg := i, arg
		promises[i] = pip.NewPip(func() V {
			printDebug("JOB START: i=%v arg=%v", i, arg)
			value, e := f(ctx, i, arg)
			if e != nil {
				once.Do(func() {
					err = e
					close(catch)
				})
			}
			printDebug("JOB DONE: i=%v arg=%v value=%v err=%v", i, arg, value, e)
			<-traffic
			return value
		})
		printDebug("promises[%v] = p", i)
		var breaker bool
		select {
		case <-catch:
			breaker = true
		case <-ctx.Done():
			breaker = true
		case traffic <- struct{}{}:
		}
		if breaker {
			promises = promises[0 : i+1]
			printDebug("SKIP %v", len(promises))
			break
		}
	}
	printDebug("LOOP END %v", len(promises))
	printDebug("close(traffic) %v", len(traffic))
	close(traffic)

	res := make([]V, len(args))
	for i, p := range promises {
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case <-catch:
			return res, err
		case msg := <-p:
			printDebug("fill %v %v %v", i, p, msg)
			res[i] = msg
		}
	}
	return res, nil
}
