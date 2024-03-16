package async

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// can save the resulting array after canceling/error: YES/YES
// throws "context canceled" if an error occurs before/after cancellation: YES/YES
// instant termination on cancelation/error: YES/YES
func AsyncErrgroup[A any, V any](c context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	printDebug("%v %v", args, concurrency)
	out := make(chan struct {
		Index int
		Value V
	}, concurrency)

	ctx, cancel := context.WithCancel(c)
	defer cancel()

	wg := new(errgroup.Group)
	traffic := make(chan struct{}, concurrency-1)

	var err error
	result := make([]V, len(args))

	go func() {
		defer func() {
			printDebug("close(traffic)")
			close(traffic)
			err = wg.Wait()
			printDebug("ERRR %v", err)
			close(out)
		}()
		for index, arg := range args {
			index := index
			arg := arg
			wg.Go(func() error {
				printDebug("wg.Go %v", arg)
				value, err := f(index, arg)
				printDebug("done: %v %v %v", arg, value, err)
				if err != nil {
					cancel()
					return err
				}
				<-traffic
				out <- struct {
					Index int
					Value V
				}{index, value}
				printDebug("return %v", arg)
				return nil
			})
			select {
			case <-ctx.Done():
				printDebug("skipping input %v %v", index, arg)
				return
			case traffic <- struct{}{}:
			}
		}
	}()

	for {
		select {
		case msg, ok := <-out:
			if !ok {
				return result, err
			}
			printDebug("msg %v", msg)
			result[msg.Index] = msg.Value
		case <-c.Done():
			return result, ctx.Err()
		}
	}
}
