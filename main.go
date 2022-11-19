package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// handle SIGINT (control+c)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		select {
		case <-c:
			fmt.Println("\nmain: interrupt received. cancelling context.")
		}
		cancel()
	}()

	rand.Seed(time.Now().UnixNano())
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
OUT:
	for i := 1; i <= 1; i++ {
		res, err := MapW(ctx, data, func(k int) (int, error) {
			rnd := rand.Intn(1000)
			<-time.After(time.Duration(rnd) * time.Millisecond)
			if rand.Intn(len(data)) == 0 {
				return k, errors.New("unknown error")
			}
			return k, nil
		}, 2)
		fmt.Printf("[%v] RESULT: %v %v\n", i, res, err)
		select {
		case <-ctx.Done():
			fmt.Println("BREAAAAAK")
			break OUT
		default:
			continue
		}
	}
	<-time.After(time.Duration(3) * time.Second)
}

func Printf(template string, rest ...interface{}) {
	args := append([]interface{}{time.Now().String()[0:25]}, rest...)
	fmt.Printf("[ %v ]    "+template+"\n", args...)
}

func MapWg[K comparable, V any](ctx context.Context, list []K, f func(k K) (V, error), concurrency int) (map[K]V, error) {
	if concurrency == 0 {
		concurrency = len(list)
	}
	Printf("CONCURRENCY: %v", concurrency)
	errchan := make(chan error)
	traffic := make(chan struct{}, concurrency-1)
	output := make(chan struct {
		Key   K
		Value V
		error
	})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}
	res := make(map[K]V, 0)

	go func() {
		for m := range output {
			Printf("m := <-output: %v %v %v", m.Key, m.Value, m.error)
			if m.error != nil {
				errchan <- m.error
				go func() {
					for range output {
						Printf("for range output")
					}
				}()
				break
			}
			res[m.Key] = m.Value
			Printf("%v", res)
		}
		Printf("---------- END OUTPUT LOOP")
		close(errchan)
	}()

	go func() {
		for _, key := range list {
			select {
			case <-ctx.Done():
				Printf("SKIP %v", key)
				continue
			default:
			}
			wg.Add(1)
			Printf("wg.Add(%v)", key)
			go func(key K) {
				Printf("go func(key K) %v", key)
				value, err := f(key)
				Printf("JOB DONE %v %v", value, err)
				output <- struct {
					Key   K
					Value V
					error
				}{key, value, err}
				Printf("wg.Done(%v)", key)
				wg.Done()
				<-traffic
			}(key)
			traffic <- struct{}{}
		}
		Printf("---------- END INPUT LOOP %v", len(traffic))
		close(traffic)

		wg.Wait()
		Printf("---------wg.Wait()")
		close(output)
	}()

	select {
	case err, ok := <-errchan:
		Printf("%v %v", err, ok)
		if err != nil {
			return nil, err
		}
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// array of channels
func MapArr[A any, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	Printf("CONCURRENCY: %v", concurrency)

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
				fmt.Println("SKIP", len(promises))
				break
			}
			p := make(chan struct {
				Index int
				Value V
				error
			}, 1)
			promises[i] = p
			Printf("promises[%v] = p", i)
			go func(i int, arg A) {
				Printf("go func() %v %v", i, arg)
				value, err := f(arg)
				Printf("JOB DONE: %v %v %v", arg, value, err)
				if err != nil {
					atomic.CompareAndSwapInt32(&stop, 0, 1)
				}
				p <- struct {
					Index int
					Value V
					error
				}{i, value, err}
				close(p)
				Printf("promises[%v] <- %v %v", i, arg, value)
				<-traffic
			}(i, arg)
			traffic <- struct{}{}
		}
		close(traffic)
		Printf("close(traffic)")

		wg.Add(len(promises))
		for i, p := range promises {
			Printf("pipe %v %v", i, p)
			i, p := i, p
			go func() {
				pipe <- <-p
				wg.Done()
				Printf("wg.Done(%v)", i)
			}()
		}
		Printf("pipe <- <-p")

		wg.Wait()
		Printf("close(pipe)")
		close(pipe)
	}()

	res := make([]V, len(args))
	for {
		select {
		case <-ctx.Done():
			fmt.Println("<-ctx.Done():")
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
			Printf("<- promises[%v] %v", m.Index, res)
		}
	}
}

// workers
func MapW[A comparable, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) (map[A]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	in, out := make(chan A), make(chan struct {
		Key   A
		Value V
		error
	}, concurrency)

	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	worker := func(i int) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				runtime.Gosched()
			}
			arg, ok := <-in
			if !ok {
				return
			}
			value, err := f(arg)
			Printf("worker %v done, res[%v] = %v, err = %v", i+1, arg, value, err)
			out <- struct {
				Key   A
				Value V
				error
			}{arg, value, err}
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}
	Printf("%v workers created", concurrency)

	go func() {
	OUT:
		for _, arg := range args {
			select {
			case <-ctx.Done():
				fmt.Println("skipping input")
				break OUT
			case in <- arg:
			}
		}
		close(in)
		Printf("input channel closed")
	}()

	go func() {
		wg.Wait()
		close(out)
		Printf("ouptut channel closed")
	}()

	result := make(map[A]V, len(args))
	for m := range out {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if m.error != nil {
				return nil, m.error
			}
			result[m.Key] = m.Value
		}
	}

	return result, nil
}
