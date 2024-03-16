package async

import (
	"context"
	"sync"
)

// can save the resulting array after canceling/error: YES/YES
// throws "context canceled" if an error occurs before/after cancellation: YES/YES
// instant termination on cancelation/error: YES/YES
func AsyncSemaphore[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)
	complete := make(chan error)
	traffic := make(chan struct{}, concurrency-1)
	output := make(chan struct {
		Index int
		Value V
		error
	}, concurrency) // size=concurrency to prevent blocking of the input channel

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}
	res := make([]V, len(args))

	go func() {
		defer func() {
			printDebug("-------- END INPUT LOOP")
			printDebug("traffic channel closed (tail %v)", len(traffic))
			close(traffic)
			printDebug("wg.Wait()")
			wg.Wait()
			printDebug("output channel closed (tail %v)", len(output))
			close(output)
		}()
		for i, arg := range args {
			select {
			case <-ctx.Done():
				printDebug("SKIP INPUT %v", arg)
				return
			default:
			}
			wg.Add(1)
			printDebug(" + wg.Add(%v)", arg)
			go func(i int, arg A) {
				printDebug("go func(%v)", arg)
				value, err := f(i, arg)
				printDebug("CHAN <- struct {%v, %v, %v}", i, value, err)
				output <- struct {
					Index int
					Value V
					error
				}{i, value, err}
				printDebug(" - wg.Done(%v)", arg)
				wg.Done()
				<-traffic
			}(i, arg)
			traffic <- struct{}{}
		}
	}()

	go func() {
		var err error
		defer func() {
			printDebug("-------- END OUTPUT LOOD %v", err)
			complete <- err
			close(complete)
		}()
		for {
			select {
			case <-ctx.Done():
				printDebug("SKIP OUTPUT %v")
				err = ctx.Err()
				return
			case msg, ok := <-output:
				printDebug("CHAN msg := struct {%v, %v, %v, %v}", msg.Index, msg.Value, msg.error, ok)
				if !ok {
					return
				}
				if msg.error != nil {
					err = msg.error
					return
				}
				res[msg.Index] = msg.Value
				printDebug("_____%v %v %v %v %v", res, msg.Index, msg.Value, msg.error)
			}
		}
	}()

	select {
	case err, ok := <-complete:
		printDebug("err, ok := %v, %v", err, ok)
		if err != nil {
			return res, err
		}
		return res, nil
		//case <-ctx.Done():
		//	// cant return "res" because of DATA RACE
		//	return nil, ctx.Err()

	}
}
