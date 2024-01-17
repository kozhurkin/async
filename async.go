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

func Pipeline[R any, C chan R](f func() R) C {
	out := make(C, 1)
	go func() {
		out <- f()
		close(out)
	}()
	return out
}

func AsyncToMap[A comparable, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) (map[A]V, error) {
	arr, err := AsyncToArray(ctx, args, f, concurrency)
	if err != nil {
		return nil, err
	}
	res := make(map[A]V, len(args))
	for i, a := range args {
		res[a] = arr[i]
	}
	return res, nil
}

func AsyncToArray[A any, V any](ctx context.Context, args []A, f func(k A) (V, error), concurrency int) ([]V, error) {
	return AsyncWorkers(ctx, args, f, concurrency)
}
