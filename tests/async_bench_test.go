package tests

import (
	"context"
	"errors"
	"github.com/kozhurkin/async"
	"math/rand"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

var datas = func() [][]int {
	res := [][]int{
		make([]int, 1),
		make([]int, 4),
		make([]int, 16),
		make([]int, 64),
		make([]int, 256),
	}
	for _, data := range res {
		for i := range data {
			data[i] = rand.Intn(10000)
		}
	}
	return res
}()

var seed = time.Now().UnixNano()

func launcher(b *testing.B, asyncFunc func(context.Context, []int, func(context.Context, int, int) (int, error), int) ([]int, error)) {
	ctx := context.Background()
	rand.Seed(seed)
	var errs, iters int32
	for _, data := range datas {
		length := len(data)
		for c := 0; c <= 10; c++ {
			for i := 1; i <= b.N; i++ {
				asyncFunc(ctx, data, func(ctx context.Context, i int, k int) (int, error) {
					rnd := rand.Intn(1000)
					runtime.Gosched()
					//<-time.After(time.Duration(rnd) * time.Nanosecond)
					atomic.AddInt32(&iters, 1)
					if rand.Intn(length) == 0 {
						atomic.AddInt32(&errs, 1)
						return rnd, errors.New("unknown error")
					}
					return rnd, nil
				}, c)
			}
		}
	}
	//fmt.Println("throw", b.N, atomic.LoadInt32(&iters), atomic.LoadInt32(&errs))
}

func BenchmarkAsyncSemaphore(b *testing.B) {
	launcher(b, async.AsyncSemaphore[int, int])
}

func BenchmarkAsyncWorkers(b *testing.B) {
	launcher(b, async.AsyncWorkers[int, int])
}

func BenchmarkAsyncPipers(b *testing.B) {
	launcher(b, async.AsyncPipers[int, int])
}

func BenchmarkAsyncPromiseAtomic(b *testing.B) {
	launcher(b, async.AsyncPromiseAtomic[int, int])
}

func BenchmarkAsyncErrgroup(b *testing.B) {
	launcher(b, async.AsyncErrgroup[int, int])
}

func BenchmarkAsyncErrgroupSimple(b *testing.B) {
	launcher(b, async.AsyncErrgroupSimple[int, int])
}

func BenchmarkAsyncPromiseCatch(b *testing.B) {
	launcher(b, async.AsyncPromiseCatch[int, int])
}

func BenchmarkAsyncPromiseSync(b *testing.B) {
	launcher(b, async.AsyncPromiseSync[int, int])
}

func BenchmarkAsyncPromisePipes(b *testing.B) {
	launcher(b, async.AsyncPromisePipes[int, int])
}
