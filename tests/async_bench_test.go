package tests

import (
	"context"
	"errors"
	"github.com/kozhurkin/async"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

var data = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
var handler = func(i int, k int) (int, error) {
	rnd := rand.Intn(10000)
	runtime.Gosched()
	//<-time.After(time.Duration(rnd) * time.Nanosecond)
	if rand.Intn(len(data)) == 0 {
		return rnd, errors.New("unknown error")
	}
	return rnd, nil
}

func BenchmarkAsyncPromisePipes(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			async.AsyncPromisePipes(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncSemaphore(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			async.AsyncSemaphore(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncPromiseAtomic(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			async.AsyncPromiseAtomic(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncPromiseCatch(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			async.AsyncPromiseCatch(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncPromiseSync(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			async.AsyncPromiseSync(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncWorkers(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			async.AsyncWorkers(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncErrgroup(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			async.AsyncErrgroup(ctx, data, handler, c)
		}
	}
}
