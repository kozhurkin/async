package async

import (
	"context"
	"github.com/kozhurkin/async/pipers"
)

// can save the resulting array after canceling/error: YES/YES
// throws "context canceled" if an error occurs before/after cancellation: YES/YES
// instant termination on cancelation/error: YES/YES
func AsyncPipers[A any, V any](ctx context.Context, args []A, f func(int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	traffic := make(chan struct{}, concurrency)
	pp := make(pipers.Pipers[V], len(args))
	stop := make(chan struct{})

	for i, a := range args {
		i, a := i, a
		pp[i] = pipers.NewPiper(func() (V, error) {
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
		for _, pp := range pp {
			select {
			case <-stop:
				pp.Close()
			case traffic <- struct{}{}:
				pp.Run()
			}
		}
	}()

	err := pp.FirstErrorContext(ctx)
	close(stop)
	printDebug("close(stop)")

	return pp.Results(), err
}
