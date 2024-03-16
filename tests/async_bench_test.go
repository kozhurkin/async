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

func launcher(b *testing.B, asyncFunc func(context.Context, []int, func(int, int) (int, error), int) ([]int, error)) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			asyncFunc(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncSemaphore(b *testing.B) {
	launcher(b, async.AsyncSemaphore[int, int])
}

func BenchmarkAsyncWorkers(b *testing.B) {
	launcher(b, async.AsyncWorkers[int, int])
}

func BenchmarkAsyncErrgroup(b *testing.B) {
	launcher(b, async.AsyncErrgroup[int, int])
}

func BenchmarkAsyncPromiseCatch(b *testing.B) {
	launcher(b, async.AsyncPromiseCatch[int, int])
}

func BenchmarkAsyncPromiseAtomic(b *testing.B) {
	launcher(b, async.AsyncPromiseAtomic[int, int])
}

func BenchmarkAsyncPromiseSync(b *testing.B) {
	launcher(b, async.AsyncPromiseSync[int, int])
}

func BenchmarkAsyncPromisePipes(b *testing.B) {
	launcher(b, async.AsyncPromisePipes[int, int])
}
