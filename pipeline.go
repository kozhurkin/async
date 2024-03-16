package async

import (
	"fmt"
	"sync"
)

type PipErr[R any] struct {
	Out chan R
	Err chan error
}

func PipelineErr[R any](f func() (R, error)) PipErr[R] {
	outch := make(chan R, 1)
	errch := make(chan error, 1)
	go func() {
		v, e := f()
		go func() {
			outch <- v
			close(outch)
		}()
		go func() {
			errch <- e
			close(errch)
		}()
	}()
	return PipErr[R]{outch, errch}
}

func PipelineErrReducer[R any](piperrs []PipErr[R]) ([]R, error) {
	size := len(piperrs)
	res := make([]R, size)
	errchan := make(chan error)
	wg := sync.WaitGroup{}

	wg.Add(size)
	for _, pipe := range piperrs {
		pipe := pipe
		go func() {
			errchan <- <-pipe.Err
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(errchan)
	}()

	for err := range errchan {
		if err != nil {
			return nil, err
		}
	}
	for i, p := range piperrs {
		fmt.Printf("%v %v %T\n", i, p, p)
		res[i] = <-p.Out
	}
	return res, nil
}

func Pipeline[R any, C chan R](f func() R) C {
	out := make(C, 1)
	go func() {
		out <- f()
		close(out)
	}()
	return out
}

func PipelineReducer[R any, C []chan R](pipes C) []R {
	res := make([]R, len(pipes))
	for i, p := range pipes {
		fmt.Printf("%v %v %T\n", i, p, p)
		res[i] = <-p
	}
	return res
}
