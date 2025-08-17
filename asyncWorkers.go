package async

import (
	"context"
	"sync"
)

// tests: ✅
// bench: ✅

func AsyncWorkers[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
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

	worker := func(w int) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		defer wg.Done()
		for input := range in {
			if ctx.Err() != nil {
				return
			}
			value, err := f(ctx, input.Index, input.Arg)
			out <- struct {
				Index int
				Value V
				error
			}{input.Index, value, err}
			if err != nil {
				cancel()
			}
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}

	go func() {
		defer close(in)
		for i, arg := range args {
			select {
			case <-ctx.Done():
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
