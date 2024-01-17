package async

import (
	"context"
	"runtime"
	"sync"
)

func AsyncSemaphore[A any, V any](ctx context.Context, args []A, f func(k A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)
	end := make(chan error)
	traffic := make(chan struct{}, concurrency-1)
	output := make(chan struct {
		Index int
		Value V
		error
	})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}
	res := make([]V, len(args))

	go func() {
	LOOP:
		for i, arg := range args {
			select {
			case <-ctx.Done():
				printDebug("SKIP %v", arg)
				// TODO break OUT?
				break LOOP
			default:
			}
			wg.Add(1)
			printDebug(" + wg.Add(%v)", arg)
			go func(i int, arg A) {
				printDebug("go func(%v)", arg)
				value, err := f(arg)
				printDebug("CHAN <- struct {%v, %v, %v}", i, value, err)
				output <- struct {
					Index int
					Value V
					error
				}{i, value, err}
				printDebug(" - wg.Done(%v)", arg)
				wg.Done()
				// switch goroutine to handle error and exit before next iteration
				// it works without it, but it saves you from unnecessarily running the task
				if err != nil {
					runtime.Gosched()
				}
				<-traffic
			}(i, arg)
			traffic <- struct{}{}
		}
		printDebug("LOOP INPUT DONE")
		printDebug("traffic channel closed (tail %v)", len(traffic))
		close(traffic)

		printDebug("wg.Wait()")
		wg.Wait()
		printDebug("output channel closed (tail %v)", len(output))
		close(output)
	}()

	go func() {
		var err error
		for msg := range output {
			printDebug("CHAN msg := struct {%v, %v, %v}", msg.Index, msg.Value, msg.error)
			if msg.error != nil {
				err = msg.error
				break
			}
			res[msg.Index] = msg.Value
			printDebug("%v", res)
		}
		printDebug("LOOP OUTPUT DONE")
		end <- err
		close(end)
	}()

	select {
	case err, ok := <-end:
		printDebug("err, ok := %v, %v", err, ok)
		if err != nil {
			return nil, err
		}
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
