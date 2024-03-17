package async

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type Piper[R any] struct {
	Out chan R
	Err chan error
}

type Pipers[R any] []Piper[R]

func (pp Pipers[R]) Results() []R {
	res := make([]R, len(pp))
	for i, p := range pp {
		select {
		case res[i] = <-p.Out:
		default:

		}
	}
	return res
}
func (pp Pipers[R]) ErrorsAll() []error {
	return pp.FirstNErrors(0)
}

func (pp Pipers[R]) ErrorsAllContext(ctx context.Context) []error {
	return pp.FirstNErrorsContext(ctx, 0)
}

func (pp Pipers[R]) FirstNErrorsContext(ctx context.Context, n int) []error {
	res := make([]error, 0, n)
	errchan := pp.ErrorsChan()
	done := make(chan struct{})
	var closedone int32
	go func() {
		if sig := <-ctx.Done(); atomic.LoadInt32(&closedone) == 0 {
			done <- sig
		}
	}()
	defer func() {
		atomic.AddInt32(&closedone, 1)
		close(done)
	}()
	for {
		select {
		case err, ok := <-errchan:
			if !ok {
				if len(res) == 0 {
					return nil
				}
				return res
			} else if err != nil {
				res = append(res, err)
			}
		case <-done:
			res = append(res, ctx.Err())
		}
		if n > 0 && len(res) == n {
			return res
		}
	}
}

func (pp Pipers[R]) FirstNErrors(n int) []error {
	return pp.FirstNErrorsContext(context.Background(), n)
}

func (pp Pipers[R]) FirstError() error {
	for err := range pp.ErrorsChan() {
		if err != nil {
			return err
		}
	}
	return nil
}

func (pp Pipers[R]) FirstErrorContext(ctx context.Context) error {
	errchan := pp.ErrorsChan()
	for {
		select {
		case err, ok := <-errchan:
			if !ok {
				return nil
			} else if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (pp Pipers[R]) ErrorsChan() chan error {
	size := len(pp)
	errchan := make(chan error, size)
	wg := sync.WaitGroup{}

	wg.Add(size)
	for _, pipe := range pp {
		pipe := pipe
		go func() {
			e := <-pipe.Err
			errchan <- e
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(errchan)
	}()

	return errchan
}

func (pp Pipers[R]) Resolve() ([]R, error) {
	err := pp.FirstError()
	return pp.Results(), err
}

func (pp Pipers[R]) ResolveContext(ctx context.Context) ([]R, error) {
	err := pp.FirstErrorContext(ctx)
	return pp.Results(), err
}

func NewPiper[R any](f func() (R, error)) Piper[R] {
	outch := make(chan R, 1)
	errch := make(chan error, 1)
	go func() {
		v, e := f()
		outch <- v
		close(outch)
		errch <- e
		close(errch)

	}()
	return Piper[R]{outch, errch}
}

func PipersResolve[R any](pipes ...Piper[R]) ([]R, error) {
	return Pipers[R](pipes).Resolve()
}

func Pipeline[R any, C chan R](f func() R) C {
	out := make(C, 1)
	go func() {
		out <- f()
		close(out)
	}()
	return out
}

func PipelineReducer[R any](pipes ...chan R) []R {
	res := make([]R, len(pipes))
	for i, p := range pipes {
		fmt.Printf("%v %v %T\n", i, p, p)
		res[i] = <-p
	}
	return res
}
