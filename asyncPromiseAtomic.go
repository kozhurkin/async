package async

import (
	"context"
	"sync"
	"sync/atomic"
)

// save the resulting array after canceling/error: SOSO/YES
// throws "context canceled" if an error occurs before/after cancellation: YES/YES
// instant termination on cancelation/error: YES/YES
func AsyncPromiseAtomic[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	promises := make([]chan struct {
		Index int
		Value V
		error
	}, len(args))

	traffic := make(chan struct{}, concurrency-1)

	output := make(chan struct {
		Index int
		Value V
		error
	}, len(args))

	var stop int32
	var wg = sync.WaitGroup{}

	go func() {
		for i, arg := range args {
			if atomic.LoadInt32(&stop) == 1 {
				promises = promises[0:i]
				printDebug("SKIP %v", len(promises))
				break
			}
			i, arg := i, arg
			promises[i] = Pipeline(func() struct {
				Index int
				Value V
				error
			} {
				printDebug("JOB START: i=%v arg=%v", i, arg)
				value, err := f(i, arg)
				printDebug("JOB DONE: i=%v arg=%v value=%v err=%v", i, arg, value, err)
				if err != nil {
					atomic.CompareAndSwapInt32(&stop, 0, 1)
				}
				printDebug("promises[%v] <- {%v %v}", i, arg, value)
				<-traffic
				return struct {
					Index int
					Value V
					error
				}{i, value, err}
			})
			printDebug("promises[%v] = p", i)
			traffic <- struct{}{}
		}
		close(traffic)
		printDebug("close(traffic) %v", len(traffic))

		wg.Add(len(promises))
		for i, p := range promises {
			printDebug("pipe %v %v", i, p)
			i, p := i, p
			go func() { // need when error exist in tail unfulfilled promises
				output <- <-p
				wg.Done()
				printDebug("wg.Done(%v)", i)
			}()
		}

		printDebug("wg.Wait()")
		wg.Wait()
		printDebug("close(output)")
		close(output)
	}()

	res := make([]V, len(args))
	for {
		select {
		case <-ctx.Done():
			printDebug("<-ctx.Done():")
			atomic.CompareAndSwapInt32(&stop, 0, 1)
			return res, ctx.Err()
		case m, ok := <-output:
			if !ok {
				return res, nil
			}
			if m.error != nil {
				return res, m.error
			}
			res[m.Index] = m.Value
			printDebug("<- promises[%v] %v", m.Index, res)
		}
	}
}
