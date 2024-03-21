package pipers

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var debug = 0

func printDebug(template string, rest ...interface{}) {
	if debug == 1 {
		args := append([]interface{}{time.Now().String()[0:25]}, rest...)
		fmt.Printf("pipers:  [ %v ]    "+template+"\n", args...)
	}
}

type Piper[R any] struct {
	Out chan R
	Err chan error
	Job func() (R, error)
}

func (p Piper[R]) Close() Piper[R] {
	printDebug(".Close(%v)", p)
	close(p.Out)
	close(p.Err)
	return p
}
func (p Piper[R]) Run() Piper[R] {
	p.run()
	return p
}
func (p Piper[R]) run() chan error {
	printDebug(".run(%v)  ", p)
	done := make(chan error, 1)
	go func() {
		v, e := p.Job()
		if e != nil {
			done <- e
		}
		close(done)
		p.Out <- v
		p.Err <- e
		p.Close()
	}()
	return done
}

func NewPiper[R any](f func() (R, error)) Piper[R] {
	return Piper[R]{
		Out: make(chan R, 1),
		Err: make(chan error, 1),
		Job: f,
	}
}

// Concurrency
// Context

type Pipers[R any] []Piper[R]

func (pp Pipers[R]) Run() Pipers[R] {
	for _, p := range pp {
		p.Run()
	}
	return pp
}

func (pp Pipers[R]) RunConcurrency(n int) Pipers[R] {
	if n == 0 {
		return pp.Run()
	}
	go func() {
		traffic := make(chan struct{}, n)
		catch := make(chan struct{})
		var once sync.Once
		defer func() {
			printDebug("close(traffic)")
			close(traffic)
			once.Do(func() {
				printDebug("close(catch)")
				close(catch)
			})
		}()
		for i, p := range pp {
			i, p := i, p
			select {
			case <-catch:
				printDebug("___p.Close(%v)", i)
				p.Close()
			case traffic <- struct{}{}:
				printDebug("___go func(%v)", i)
				go func() {
					if err := <-p.run(); err != nil {
						once.Do(func() {
							close(catch)
						})
					} else {
						<-traffic
					}
				}()
			}
		}
	}()
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
			}
			res = append(res, err)
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
	return <-pp.ErrorsChan()
}

func (pp Pipers[R]) FirstErrorContext(ctx context.Context) error {
	errchan := pp.ErrorsChan()
	for {
		select {
		case err, ok := <-errchan:
			if !ok {
				return nil
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (pp Pipers[R]) ErrorsChan() chan error {
	errchan := make(chan error, len(pp))
	wg := sync.WaitGroup{}

	wg.Add(len(pp))
	for _, p := range pp {
		p := p
		go func() {
			if err := <-p.Err; err != nil {
				errchan <- err
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		printDebug("************** close(errchan)")
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

func NewPipers[R any](funcs ...func() (R, error)) Pipers[R] {
	res := make(Pipers[R], len(funcs))
	for i, f := range funcs {
		res[i] = NewPiper(f)
	}
	return res
}

func NewPipersMap[R any](input []R, f func(int, R) (R, error)) Pipers[R] {
	res := make(Pipers[R], len(input))
	for i, v := range input {
		i, v := i, v
		res[i] = NewPiper(func() (R, error) {
			return f(i, v)
		})
	}
	return res
}

func Ref[I any](p *I, f func() (I, error)) func() (interface{}, error) {
	return func() (interface{}, error) {
		res, err := f()
		*p = res
		return res, err
	}
}
