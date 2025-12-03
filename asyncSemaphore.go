package async

import (
	"context"
	"sync"
)

// tests: ✅
// bench: ✅
func AsyncSemaphore[A any, V any](ctx context.Context, args []A, fn func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	traffic := make(chan struct{}, concurrency)
	output := make(chan struct {
		Index int
		Value V
		error
	}, concurrency) // size=concurrency to prevent blocking of the input channel

	wg := sync.WaitGroup{}
	res := make([]V, len(args))

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		defer func() {
			//close(traffic) // not needed
			wg.Wait()
			close(output)
		}()

		for i, arg := range args {
			i, arg := i, arg

			select {
			case traffic <- struct{}{}:
			case <-ctx.Done():
				return
			}

			wg.Add(1)
			go func() {
				defer func() {
					<-traffic // освободить слот
					wg.Done() // затем отметить завершение
				}()
				value, err := fn(ctx, i, arg)
				output <- struct {
					Index int
					Value V
					error
				}{i, value, err}
				if err != nil {
					cancel()
				}
			}()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case msg, ok := <-output:
			if !ok {
				return res, nil
			}
			if msg.error != nil {
				return res, msg.error
			}
			res[msg.Index] = msg.Value
		}
	}
}
