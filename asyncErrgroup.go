package async

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// errgroup
func AsyncErrgroup[A any, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	printDebug("%v %v", args, concurrency)
	out := make(chan struct {
		Index int
		Value V
	}, concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := new(errgroup.Group)
	traffic := make(chan struct{}, concurrency-1)

	var err error
	result := make([]V, len(args))

	go func() {
		for index, arg := range args {
			index := index
			arg := arg
			select {
			case <-ctx.Done():
				printDebug("skipping input %v %v", index, arg)
				continue
			default:
			}
			wg.Go(func() error {
				printDebug("wg.Go %v", arg)
				value, err := f(arg)
				printDebug("done: %v %v %v", arg, value, err)
				go func() {
					<-traffic
				}()
				if err != nil {
					cancel()
					return err
				}
				out <- struct {
					Index int
					Value V
				}{index, value}
				printDebug("return %v", arg)
				return nil
			})
			traffic <- struct{}{}
		}
		printDebug("close(traffic)")
		close(traffic)

		err = wg.Wait()
		printDebug("ERRR %v", err)
		close(out)
	}()

	for msg := range out {
		printDebug("msg %v", msg)
		result[msg.Index] = msg.Value
	}

	select {
	case <-ctx.Done():
		if err == nil {
			err = ctx.Err()
		}
	default:
	}

	return result, err
}
