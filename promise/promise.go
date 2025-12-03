package promise

import (
	"sync"
)

type Promise[R any] struct {
	job    func() (R, error)
	Out    chan *R
	Err    chan error
	runned chan struct{}
	done   chan struct{}
	result *R
	err    error
	once   sync.Once
	mu     sync.Mutex
}

func (p *Promise[R]) markAsRunned() {
	close(p.runned)
}

func (p *Promise[R]) markAsDone() {
	close(p.done)
}

func (p *Promise[R]) Done() <-chan struct{} {
	return p.done
}

func (p *Promise[R]) Completed() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

func (p *Promise[R]) Runned() bool {
	select {
	case <-p.runned:
		return true
	default:
		return false
	}
}

func (p *Promise[R]) Value() R {
	p.Run()

	p.mu.Lock()
	defer p.mu.Unlock()

	value, ok := <-p.Out
	if ok {
		p.result = value
		return *value
	}
	return *p.result
}

func (p *Promise[R]) Error() error {
	p.Run()

	p.mu.Lock()
	defer p.mu.Unlock()

	err, ok := <-p.Err
	if ok {
		p.err = err
		return err
	}
	return p.err
}

func (p *Promise[R]) Return() (R, error) {
	err := p.Error()
	return p.Value(), err
}

func (p *Promise[R]) Then(fn func(R) (R, error)) *Promise[R] {
	return NewPromise(func() (R, error) {
		err := p.Error()
		value := p.Value()
		if err != nil {
			return value, err
		}
		return fn(value)
	})
}

func (p *Promise[R]) Catch(fn func(error) (R, error)) *Promise[R] {
	return NewPromise(func() (R, error) {
		if err := p.Error(); err != nil {
			return fn(err)
		}
		return p.Value(), nil
	})
}

func (p *Promise[R]) Run() *Promise[R] {
	p.once.Do(func() {
		p.markAsRunned()
		go func() {
			v, e := p.job()
			p.Err <- e
			p.Out <- &v
			close(p.Out)
			close(p.Err)
			p.markAsDone()
		}()
	})
	return p
}

func Resolve[R any](v R) *Promise[R] {
	return NewPromise(func() (R, error) {
		return v, nil
	})
}
func NewPromise[R any](job func() (R, error)) *Promise[R] {
	res := Promise[R]{
		job:    job,
		Out:    make(chan *R, 1),
		Err:    make(chan error, 1),
		runned: make(chan struct{}),
		done:   make(chan struct{}),
	}
	return &res
}
