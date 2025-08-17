package async

import (
	"context"
	"sync"
)

// tests: ✅
// bench: ✅

func AsyncSemaphore[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)

	traffic := make(chan struct{}, concurrency)
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
			i, arg := i, arg

			traffic <- struct{}{}
			if ctx.Err() != nil {
				printDebug("SKIP INPUT %v", arg)
				return
			}
			wg.Add(1)
			printDebug(" + wg.Add(%v)", arg)
			go func() {
				defer wg.Done()
				defer func() { <-traffic }()
				printDebug("go func(%v)", arg)
				value, err := f(ctx, i, arg)
				printDebug("CHAN <- struct {%v, %v, %v}", i, value, err)
				output <- struct {
					Index int
					Value V
					error
				}{i, value, err}
				printDebug(" - wg.Done(%v)", arg)
				if err != nil {
					cancel()
				}
			}()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			printDebug("SKIP OUTPUT %v")
			return res, ctx.Err()
		case msg, ok := <-output:
			printDebug("CHAN msg := struct {%v, %v, %v, %v}", msg.Index, msg.Value, msg.error, ok)
			if !ok {
				return res, nil
			}
			if msg.error != nil {
				return res, msg.error
			}
			res[msg.Index] = msg.Value
			printDebug("_____%v %v %v %v", res, msg.Index, msg.Value, msg.error)
		}
	}
}
