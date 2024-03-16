package async

import (
	"context"
	"sync"
)

// can save the resulting array after canceling/error: YES/YES
// throws "context canceled" if an error occurs before/after cancellation: YES/YES
// instant termination on cancelation/error: YES/YES
func AsyncWorkers[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	worker := func(w int) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case input, ok := <-in:
				if !ok {
					return
				}
				printDebug("f(input.Arg) %v", input)
				value, err := f(input.Index, input.Arg)
				printDebug("worker %v done, res[%v] = f(%v) = %v, err = %v", w+1, input.Index, input.Arg, value, err)
				out <- struct {
					Index int
					Value V
					error
				}{input.Index, value, err}
			}
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}
	printDebug("%v workers created", concurrency)

	go func() {
		defer func() {
			close(in)
			printDebug("input channel closed")
		}()
		for i, arg := range args {
			select {
			case <-ctx.Done():
				printDebug("skipping input")
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
		printDebug("ouptut channel closed")
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
