package main

import (
	"context"
	"errors"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

var data = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
var handler = func(k int) (int, error) {
	rnd := rand.Intn(10)
	runtime.Gosched()
	//<-time.After(time.Duration(rnd) * time.Nanosecond)
	if rand.Intn(len(data)) == 0 {
		return rnd, errors.New("unknown error")
	}
	return rnd, nil
}

func BenchmarkMapChan(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			MapChan(ctx, data, handler, c)
		}
	}
}

func BenchmarkMapPromise(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			MapPromise(ctx, data, handler, c)
		}
	}
}

func BenchmarkMapPromise2(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			MapPromise2(ctx, data, handler, c)
		}
	}
}

func BenchmarkMapWorkers(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			MapWorkers(ctx, data, handler, c)
		}
	}
}
