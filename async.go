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

func AsyncToArray[A any, V any](ctx context.Context, concurrency int, args []A, f func(context.Context, int, A) (V, error)) ([]V, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return AsyncWorkers(ctx, args, f, concurrency)
}

func AsyncToMap[A comparable, V any](ctx context.Context, concurrency int, args []A, f func(context.Context, int, A) (V, error)) (map[A]V, error) {
	arr, err := AsyncToArray(ctx, concurrency, args, f)
	res := make(map[A]V, len(args))
	for i, a := range args {
		res[a] = arr[i]
	}
	return res, err
}

func AsyncFuncs[V any, F func(context.Context) (V, error)](ctx context.Context, concurrency int, funcs ...F) ([]V, error) {
	return AsyncToArray(ctx, concurrency, funcs, func(ctx context.Context, i int, f F) (V, error) {
		return f(ctx)
	})
}
