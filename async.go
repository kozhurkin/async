package async

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var debug = 0

func SetDebug(d int) {
	debug = d
}
func printDebug(template string, rest ...interface{}) {
	if debug == 1 {
		args := append([]interface{}{time.Now().String()[0:25]}, rest...)
		fmt.Printf("async:  [ %v ]    "+template+"\n", args...)
	}
}

func Pipeline[R any, C chan R](f func() R) C {
	out := make(C, 1)
	go func() {
		out <- f()
		close(out)
	}()
	return out
}

func AsyncToMap[A comparable, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) (map[A]V, error) {
	arr, err := AsyncToArray(ctx, args, f, concurrency)
	if err != nil {
		return nil, err
	}
	res := make(map[A]V, len(args))
	for i, a := range args {
		res[a] = arr[i]
	}
	return res, nil
}

func AsyncToArray[A any, V any](ctx context.Context, args []A, f func(k A) (V, error), concurrency int) ([]V, error) {
	return AsyncWorkers(ctx, args, f, concurrency)
}

func AsyncSemaphore[A any, V any](ctx context.Context, args []A, f func(k A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)
	errchan := make(chan error)
	traffic := make(chan struct{}, concurrency-1)
	output := make(chan struct {
		Index int
		Value V
		error
	})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}
	res := make([]V, len(args))

	go func() {
		for m := range output {
			printDebug("m := <-output: %v %v %v", m.Index, m.Value, m.error)
			if m.error != nil {
				errchan <- m.error
				go func() {
					for range output {
						printDebug("for range output")
					}
				}()
				break
			}
			res[m.Index] = m.Value
			printDebug("%v", res)
		}
		printDebug("---------- END OUTPUT LOOP")
		close(errchan)
	}()

	go func() {
		for i, arg := range args {
			select {
			case <-ctx.Done():
				printDebug("SKIP %v", arg)
				continue
			default:
			}
			wg.Add(1)
			printDebug("wg.Add(%v)", arg)
			go func(i int, arg A) {
				printDebug("go func(key K) %v", arg)
				value, err := f(arg)
				printDebug("JOB DONE %v %v", value, err)
				output <- struct {
					Index int
					Value V
					error
				}{i, value, err}
				printDebug("wg.Done(%v)", arg)
				wg.Done()
				<-traffic
			}(i, arg)
			traffic <- struct{}{}
		}
		printDebug("---------- END INPUT LOOP %v", len(traffic))
		close(traffic)

		wg.Wait()
		printDebug("---------wg.Wait()")
		close(output)
	}()

	select {
	case err, ok := <-errchan:
		printDebug("%v %v", err, ok)
		if err != nil {
			return nil, err
		}
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// array of channels
func AsyncPromise[A any, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	promises := make([]chan struct {
		Index int
		Value V
		error
	}, len(args))

	traffic := make(chan struct{}, concurrency-1)

	pipe := make(chan struct {
		Index int
		Value V
		error
	}, len(args))

	var stop int32
	var wg = sync.WaitGroup{}

	go func() {
		for i, arg := range args {
			if atomic.LoadInt32(&stop) == 1 {
				promises = promises[0:i]
				printDebug("SKIP %v", len(promises))
				break
			}
			i, arg := i, arg
			promises[i] = Pipeline(func() struct {
				Index int
				Value V
				error
			} {
				printDebug("JOB START: i=%v arg=%v", i, arg)
				value, err := f(arg)
				printDebug("JOB DONE: i=%v arg=%v value=%v err=%v", i, arg, value, err)
				if err != nil {
					atomic.CompareAndSwapInt32(&stop, 0, 1)
				}
				printDebug("promises[%v] <- {%v %v %v}", i, arg, value)
				<-traffic
				return struct {
					Index int
					Value V
					error
				}{i, value, err}
			})
			printDebug("promises[%v] = p", i)
			traffic <- struct{}{}
		}
		close(traffic)
		printDebug("close(traffic)")

		wg.Add(len(promises))
		for i, p := range promises {
			printDebug("pipe %v %v", i, p)
			i, p := i, p
			go func() {
				pipe <- <-p
				wg.Done()
				printDebug("wg.Done(%v)", i)
			}()
		}
		printDebug("pipe <- <-p")

		wg.Wait()
		printDebug("close(pipe)")
		close(pipe)
	}()

	res := make([]V, len(args))
	for {
		select {
		case <-ctx.Done():
			printDebug("<-ctx.Done():")
			atomic.CompareAndSwapInt32(&stop, 0, 1)
			return nil, ctx.Err()
		case m, ok := <-pipe:
			if !ok {
				return res, nil
			}
			if m.error != nil {
				return nil, m.error
			}
			res[m.Index] = m.Value
			printDebug("<- promises[%v] %v", m.Index, res)
		}
	}
}

// workers
func AsyncWorkers[A any, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	in, out := make(chan struct {
		Index int
		Arg   A
	}), make(chan struct {
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
			default:
				runtime.Gosched()
			}
			input, ok := <-in
			if !ok {
				return
			}
			value, err := f(input.Arg)
			printDebug("worker %v done, res[%v] = %v, err = %v", w+1, input.Arg, value, err)
			out <- struct {
				Index int
				Value V
				error
			}{input.Index, value, err}
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}
	printDebug("%v workers created", concurrency)

	go func() {
	OUT:
		for i, arg := range args {
			select {
			case <-ctx.Done():
				printDebug("skipping input")
				break OUT
			case in <- struct {
				Index int
				Arg   A
			}{i, arg}:
			}
		}
		close(in)
		printDebug("input channel closed")
	}()

	go func() {
		wg.Wait()
		close(out)
		printDebug("ouptut channel closed")
	}()

	result := make([]V, len(args))
	for m := range out {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if m.error != nil {
				return nil, m.error
			}
			result[m.Index] = m.Value
		}
	}

	return result, nil
}
