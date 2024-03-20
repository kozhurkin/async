package async

import (
	"context"
	pipers2 "github.com/kozhurkin/async/pipers"
)

// can save the resulting array after canceling/error: YES/YES
// throws "context canceled" if an error occurs before/after cancellation: YES/YES
// instant termination on cancelation/error: YES/YES
func AsyncPipers[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	traffic := make(chan struct{}, concurrency)
	pipers := make(pipers2.Pipers[V], len(args))
	stop := make(chan struct{})

	for i, a := range args {
		i, a := i, a
		pipers[i] = pipers2.NewPiper(func() (V, error) {
			defer func() {
				printDebug("--- i=%v, a=%v", i, a)
				<-traffic
			}()
			printDebug("+++ i=%v, a=%v", i, a)
			return f(i, a)
		})
	}

	go func() {
		defer func() {
			printDebug("close(traffic)")
			defer close(traffic)
		}()
		for _, pp := range pipers {
			select {
			case <-stop:
				pp.Close()
			case traffic <- struct{}{}:
				pp.Run()
			}
		}
	}()

	err := pipers.FirstErrorContext(ctx)
	close(stop)
	printDebug("close(stop)")

	return pipers.Results(), err
}
