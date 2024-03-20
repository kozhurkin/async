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
	Job func() (R, error)
}

func (p Piper[R]) Close() Piper[R] {
	fmt.Println("close", p)
	close(p.Out)
	close(p.Err)
	return p
}
func (p Piper[R]) Run() Piper[R] {
	fmt.Println("run  ", p)
	go func() {
		v, e := p.Job()
		p.Out <- v
		p.Err <- e
		p.Close()
	}()
	return p
}

type Pipers[R any] []Piper[R]

func (pp Pipers[R]) Run() Pipers[R] {
	for _, p := range pp {
		p.Run()
	}
	return pp
}

type Results[R any] []R

func (r *Results[R]) Shift() R {
	value := (*r)[0]
	*r = (*r)[1:len(*r)]
	return value
}

func (pp Pipers[R]) Results() Results[R] {
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
	for _, p := range pp {
		pipe := p
		go func() {
			e := <-pipe.Err
			errchan <- e
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		fmt.Println("************** close(errchan)")
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
	return Piper[R]{
		Out: make(chan R, 1),
		Err: make(chan error, 1),
		Job: f,
	}
}

func NewPipers[R any](funcs ...func() (R, error)) Pipers[R] {
	res := make(Pipers[R], len(funcs))
	for i, f := range funcs {
		res[i] = NewPiper(f)
	}
	return res
}

func PipersResolve[R any](pipes ...Piper[R]) ([]R, error) {
	return Pipers[R](pipes).Resolve()
}

func Ref[I any](p *I, f func() (I, error)) func() (interface{}, error) {
	return func() (interface{}, error) {
		res, err := f()
		*p = res
		return res, err
	}
}
