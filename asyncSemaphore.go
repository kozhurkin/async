package async

import (
	"context"
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

	res := make([]V, len(args))

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		defer func() {
			//close(traffic) // not needed
			for i := 0; i < concurrency; i++ {
				traffic <- struct{}{}
			}
			close(output)
		}()

		for i, arg := range args {
			i, arg := i, arg

			select {
			case traffic <- struct{}{}: // занять слот
			case <-ctx.Done():
				return
			}

			go func() {
				defer func() {
					<-traffic // освободить слот
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
