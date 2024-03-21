package pipers

import (
	"context"
)

type PiperSolver[R any] struct {
	Pipers[R]
	concurrency int
	ctx         context.Context
}

func (ps *PiperSolver[R]) Add(p Piper[R]) *PiperSolver[R] {
	ps.Pipers = append(ps.Pipers, p)
	return ps
}

func (ps *PiperSolver[R]) AddFunc(f func() (R, error)) *PiperSolver[R] {
	p := NewPiper(f)
	return ps.Add(p)
}

func (ps *PiperSolver[R]) Concurrency(concurrency int) *PiperSolver[R] {
	ps.concurrency = concurrency
	return ps
}

func (ps *PiperSolver[R]) Context(ctx context.Context) *PiperSolver[R] {
	ps.ctx = ctx
	return ps
}

func (ps *PiperSolver[R]) Run() *PiperSolver[R] {
	if ps.ctx == nil {
		ps.Pipers.RunContextConcurrency(context.Background(), ps.concurrency)
	} else {
		ps.Pipers.RunContextConcurrency(ps.ctx, ps.concurrency)
	}

	return ps
}

func (ps *PiperSolver[R]) FirstError() error {
	if ps.ctx == nil {
		return ps.Pipers.FirstErrorContext(context.Background())
	} else {
		return ps.Pipers.FirstErrorContext(ps.ctx)
	}
}

func (ps *PiperSolver[R]) Resolve() ([]R, error) {
	err := ps.FirstError()
	return ps.Results(), err
}
