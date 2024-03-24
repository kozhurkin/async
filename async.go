package async

import (
	"context"
	"fmt"
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

func Slice[A any, V any](ctx context.Context, concurrency int, args []A, f func(int, A) (V, error)) ([]V, error) {
	return AsyncWorkers(ctx, args, f, concurrency)
}

func SliceMapped[A comparable, V any](ctx context.Context, concurrency int, args []A, f func(int, A) (V, error)) (map[A]V, error) {
	arr, err := Slice(ctx, concurrency, args, f)
	res := make(map[A]V, len(args))
	for i, a := range args {
		res[a] = arr[i]
	}
	return res, err
}

func Funcs[V any, F func() (V, error)](ctx context.Context, concurrency int, funcs ...F) ([]V, error) {
	return AsyncWorkers(ctx, funcs, func(i int, f F) (V, error) {
		return f()
	}, concurrency)
}
