package async

import (
	"context"
	"errors"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	ts := time.Now()
	pa := Pipeline(func() int { <-time.After(1 * time.Second); return 1 })
	pb := Pipeline(func() int { <-time.After(2 * time.Second); return 2 })
	a, b := <-pa, <-pb
	delta := int(time.Now().Sub(ts).Seconds())
	if delta != 2 {
		t.Fatal("Should complete in 2 seconds")
	}
	if a != 1 || b != 2 {
		t.Fatal("Wrong return values")
	}
	return
}

var data = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
var handler = func(k int) (int, error) {
	rnd := rand.Intn(10000)
	runtime.Gosched()
	//<-time.After(time.Duration(rnd) * time.Nanosecond)
	if rand.Intn(len(data)) == 0 {
		return rnd, errors.New("unknown error")
	}
	return rnd, nil
}

func BenchmarkAsyncSemaphore(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			AsyncSemaphore(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncPromise(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			AsyncPromise(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncPromise2(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			AsyncPromise2(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncWorkers(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			AsyncWorkers(ctx, data, handler, c)
		}
	}
}

func BenchmarkAsyncErrgroup(b *testing.B) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	for c := 0; c <= len(data); c++ {
		for i := 1; i <= b.N; i++ {
			AsyncErrgroup(ctx, data, handler, c)
		}
	}
}
