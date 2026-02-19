package async

import (
	"context"
)

// tests: ✅
// bench: ✅
func AsyncWorkerPool[A any, V any](ctx context.Context, args []A, fn func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	out := make(chan struct {
		Index int
		Value V
		error
	}, concurrency)

	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	pool := NewPool(concurrency, func(index int, arg A) {
		if workCtx.Err() != nil {
			return
		}
		value, err := fn(workCtx, index, arg)
		msg := struct {
			Index int
			Value V
			error
		}{
			Index: index,
			Value: value,
			error: err,
		}
		select {
		case out <- msg:
		case <-workCtx.Done():
		}
		if err != nil {
			cancel()
		}
	})

	go func() {
		for i, arg := range args {
			if workCtx.Err() != nil {
				break
			}
			pool.Handle(i, arg)
		}
		pool.Wait()
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

type Worker struct {
	id        int
	processed int
}

type Pool[T any] struct {
	fn    func(int, T)
	queue chan *Worker
}

func NewPool[T any](n int, fn func(int, T)) *Pool[T] {
	pool := Pool[T]{
		fn:    fn,
		queue: make(chan *Worker, n),
	}
	for i := 0; i < n; i++ {
		pool.queue <- &Worker{
			id:        i,
			processed: 0,
		}
	}
	return &pool
}

func (p *Pool[T]) Handle(index int, arg T) {
	w := <-p.queue
	go func() {
		p.fn(index, arg)
		w.processed++
		p.queue <- w
	}()
}

func (p *Pool[T]) Wait() []*Worker {
	workers := make([]*Worker, cap(p.queue))
	for i := 0; i < len(workers); i++ {
		workers[i] = <-p.queue
	}
	return workers
}
