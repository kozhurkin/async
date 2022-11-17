package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ctx, cancel = context.WithCancel(ctx)

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
	for i := 1; i <= 1000; i++ {
		res, err := Map2(ctx, data, func(k int) (int, error) {
			rnd := rand.Intn(1000)
			if rand.Intn(len(data)) == 0 {
				return k, errors.New("unknown error")
			}
			<-time.After(time.Duration(rnd) * time.Millisecond)
			Printf("DONE %v", k)
			return k, nil
		}, 2)
		fmt.Printf("[%v] RESULT: %v %v\n", i, res, err)
		select {
		case <-ctx.Done():
			fmt.Println("BREAAAAAK")
			return
		default:
			continue
		}
	}
	<-time.After(time.Second)
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
	wg := sync.WaitGroup{}
	res := make(map[K]V, 0)

	wg.Add(len(list))
	go func() {
		wg.Wait()
		fmt.Println("---------wg.Wait()")
		close(output)
	}()

	go func() {
		for m := range output {
			Printf("m := <-output: %v %v %v", m.Key, m.Value, m.error)
			if m.error != nil {
				errchan <- m.error
				cancel()
				for range traffic {
					Printf("for range traffic")
				}
				break
			}
			res[m.Key] = m.Value
			Printf("%v %v", res, len(traffic))
			<-traffic
		}
		Printf("---------- END OUTPUT LOOP")
		close(errchan)
	}()

	go func() {
		for _, key := range list {
			select {
			case <-ctx.Done():
				fmt.Println("SKIP", key)
				wg.Done()
				continue
			default:
			}
			go func(key K) {
				Printf("go func(key K) %v", key)
				value, err := f(key)
				Printf("VALUE %v %v", value, err)
				output <- struct {
					Key   K
					Value V
					error
				}{key, value, err}
				wg.Done()
			}(key)
			Printf("traffic <- struct{}{} ... %v", len(traffic))
			traffic <- struct{}{}
			Printf("traffic <- struct{}{}")

		}
		Printf("---------- END INPUT LOOP %v", len(traffic))
		close(traffic)
	}()

	select {
	case err, ok := <-errchan:
		Printf("%v %v", err, ok)
		if err != nil {
			return nil, err
		}
		return res, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// array of channels
func Map2[K comparable, V any](ctx context.Context, list []K, f func(k K) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(list)
	}
	Printf("CONCURRENCY: %v", concurrency)

	promises := make([]chan struct {
		Value V
		error
	}, len(list))
	errchan := make(chan error)
	defer close(errchan)
	traffic := make(chan struct{}, concurrency)
	delta, total := time.Now(), time.Now()

	for i, key := range list {
		traffic <- struct{}{}
		i := i
		key := key
		promises[i] = make(chan struct {
			Value V
			error
		}, 1)
		go func() {
			value, err := f(key)
			promises[i] <- struct {
				Value V
				error
			}{value, err}
			close(promises[i])
			Printf("promises[%v] <- %v %v (%vms)", i, key, value, time.Now().Sub(delta))
			<-traffic
		}()
	}

	res := make([]V, len(list))

	for i := range list {
		m := <-promises[i]
		if m.error != nil {
			return nil, m.error
		}
		res[i] = m.Value
		//<-traffic
		Printf("<- promises[%v] %v", i, res[i])
	}

	fmt.Println("", res, time.Now().Sub(total))

	return res, nil
}
