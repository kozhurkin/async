package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func main() {

	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	res, err := Map(data, func(k int) (int, error) {
		rnd := rand.Intn(1000)
		if rnd > 500 {
			return rnd, errors.New("unknown error")
		}
		<-time.After(time.Duration(1000) * time.Millisecond)
		return rnd, nil
	}, 1)
	<-time.After(500)
	fmt.Println("RESULT:", res, err)
}

func Map[K comparable, V any](list []K, f func(k K) (V, error), concurrency int) (map[K]V, error) {
	if concurrency == 0 {
		concurrency = len(list)
	}
	concurrency -= 1
	traffic := make(chan struct{}, concurrency)
	defer close(traffic)

	ch := make(chan struct {
		Key   K
		Value V
		error
	})
	defer close(ch)

	delta, total := time.Now(), time.Now()

	for _, key := range list {
		go func(key K) {
			traffic <- struct{}{}
			fmt.Println("   traffic <-\n")
			value, err := f(key)
			fmt.Printf("ch <- %v [Value=%v, Error=%v] (%v)\n", key, value, err, time.Now().Sub(delta))
			ch <- struct {
				Key   K
				Value V
				error
			}{key, value, err}
		}(key)
	}

	res := make(map[K]V, 0)
	for range list {
		lock := <-traffic
		fmt.Println("<- traffic", lock)
		m := <-ch
		fmt.Printf("<- ch %v [Value=%v, Error=%v]\n", m.Key, m.Value, m.error)
		if m.error != nil {
			fmt.Printf("%v [Value=%v, Error=%v] (%s)\n", m.error, m.Key, m.Value, time.Now().Sub(total))
			fmt.Printf("len(traffic)=%v\n", len(traffic))
			return nil, m.error
		}
		res[m.Key] = m.Value
	}

	fmt.Println(res, time.Now().Sub(total))
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
			fmt.Printf("promises[%v] <- %v (%vms)\n", i, value, time.Now().Sub(delta))
			close(promises[i])
		}(i, key)
	}

	res := make([]V, len(list))

	for i := range list {
		v := <-promises[i]
		fmt.Printf("<- promises[%v] %v\n", i, v)
		res[i] = v
	}

	fmt.Println(res, time.Now().Sub(total))

	return res, nil
}
