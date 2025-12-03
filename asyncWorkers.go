package async

import (
	"context"
	"sync"
)

// tests: ✅
// bench: ✅
func AsyncWorkers[A any, V any](ctx context.Context, args []A, fn func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	in := make(chan struct {
		Index int
		Arg   A
	})
	out := make(chan struct {
		Index int
		Value V
		error
	}, concurrency)

	wg := sync.WaitGroup{}

	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	worker := func(w int) {
		defer wg.Done()
		for input := range in {
			if workCtx.Err() != nil {
				return
			}
			value, err := fn(workCtx, input.Index, input.Arg)
			msg := struct {
				Index int
				Value V
				error
			}{input.Index, value, err}
			select {
			case out <- msg:
			case <-workCtx.Done():
			}
			if err != nil {
				cancel()
			}
		}
	}

	// start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}

	// send jobs
	go func() {
		defer close(in)
		for i, arg := range args {
			select {
			case <-workCtx.Done():
				return
			case in <- struct {
				Index int
				Arg   A
			}{i, arg}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	result := make([]V, len(args))
	for {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case m, ok := <-out:
			if !ok {
				return result, nil
			}
			if m.error != nil {
				return result, m.error
			}
			result[m.Index] = m.Value
		}
	}
}
