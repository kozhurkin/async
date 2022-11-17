package main

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	res, err := Map(data, func(k int) (int, error) {
		rnd := rand.Intn(1000)
		if rnd > 900 {
			return rnd, errors.New("unknown error")
		}
		<-time.After(time.Duration(rnd) * time.Millisecond)
		Printf("DONE %v", k)
		return rnd, nil
	}, 3)
	<-time.After(time.Second)
	fmt.Println("RESULT:", res, err)
}

func Printf(template string, rest ...interface{}) {
	args := append([]interface{}{time.Now()}, rest...)
	fmt.Printf("[%v]    "+template+"\n", args...)
}

func Map[K comparable, V any](list []K, f func(k K) (V, error), concurrency int) (map[K]V, error) {
	if concurrency == 0 {
		concurrency = len(list)
	}
	fmt.Println("CONCURRENCY:", concurrency)
	errchan := make(chan error)
	traffic := make(chan struct{}, concurrency-1)
	outchan := make(chan struct {
		Key   K
		Value V
		error
	})

	var stop int32
	res := make(map[K]V, 0)

	go func() {
		defer close(errchan)
		defer close(outchan)
		for range list {
			m := <-outchan
			Printf("m := <-outchan: %v %v %v", m.Key, m.Value, m.error)
			if m.error != nil {
				errchan <- m.error
				break
			}
			res[m.Key] = m.Value
			fmt.Println(res, len(traffic))
			<-traffic
		}
	}()

	go func() {
		for _, key := range list {
			go func(key K) {
				Printf("go func(key K) %v", key)
				value, err := f(key)
				Printf("VALUE %v %v", value, err)
				if atomic.LoadInt32(&stop) != 0 {
					return
				}
				outchan <- struct {
					Key   K
					Value V
					error
				}{key, value, err}
			}(key)
			Printf("traffic <- struct{}{} ... %v", len(traffic))
			traffic <- struct{}{}
			Printf("traffic <- struct{}{}")
		}
		fmt.Println("____________Dasdasd", len(traffic))
		close(traffic)
	}()

	err := <-errchan

	atomic.AddInt32(&stop, 1)

	return res, err
}

func Map_[K comparable, V any](list []K, f func(k K) (V, error), concurrency int) (map[K]V, error) {
	if concurrency == 0 {
		concurrency = len(list)
	}
	fmt.Println("CONCURRENCY:", concurrency)
	traffic := make(chan K, concurrency-1)
	ch := make(chan struct {
		Key   K
		Value V
		error
	}, concurrency)
	defer close(ch)

	delta, total := time.Now(), time.Now()

	var stop int32

	go func() {
		for _, key := range list {
			Printf("... traffic <- %v", key)
			traffic <- key
			Printf("    traffic <- %v", key)
			if atomic.LoadInt32(&stop) != 0 {
				break
			}
			go func(key K) {
				Printf("go func(key K) %v", key)
				value, err := f(key)
				Printf("... ch <- %v [Value=%v, Error=%v] (%v)", key, value, err, time.Now().Sub(delta))
				if atomic.LoadInt32(&stop) != 0 {
					return
				}
				Printf("11111111111 %v %v", key, atomic.LoadInt32(&stop))
				ch <- struct {
					Key   K
					Value V
					error
				}{key, value, err}
				Printf("    ch <- %v [Value=%v, Error=%v] (%v)", key, value, err, time.Now().Sub(delta))
			}(key)
		}
		close(traffic)
		Printf("CLOOSE(CH)")
	}()

	res := make(map[K]V, 0)
	for range list {
		Printf(" <- traffic ...")
		lock := <-traffic
		Printf(" <- traffic %v", lock)
		Printf(" <- ch ...")
		m := <-ch
		Printf(" <- ch %v [Value=%v, Error=%v]", m.Key, m.Value, m.error)
		if m.error != nil {
			Printf("ERROR: %v [Key=%v, Value=%v] (%s)", m.error, m.Key, m.Value, time.Now().Sub(total))
			Printf("len(traffic)=%v, NumGoroutine=%v", len(traffic), runtime.NumGoroutine())
			atomic.AddInt32(&stop, 1)
			Printf("STOP")
			for range traffic {
				Printf("%v %v", len(traffic), cap(traffic))
			}
			return nil, m.error
		}
		res[m.Key] = m.Value
	}

	fmt.Println("RETURN:", res, time.Now().Sub(total))
	return res, nil
}

// array of channels
func Map2[K comparable, V any](list []K, f func(k K) (V, error)) ([]V, error) {
	promises := make([]chan V, len(list))
	delta, total := time.Now(), time.Now()

	for i, key := range list {
		promises[i] = make(chan V, 1)
		go func(i int, key K) {
			value, _ := f(key)
			promises[i] <- value
			Printf("promises[%v] <- %v (%vms)", i, value, time.Now().Sub(delta))
			close(promises[i])
		}(i, key)
	}

	res := make([]V, len(list))

	for i := range list {
		v := <-promises[i]
		Printf("<- promises[%v] %v", i, v)
		res[i] = v
	}

	fmt.Println(res, time.Now().Sub(total))

	return res, nil
}
