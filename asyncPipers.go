package async

import (
	"context"
	"github.com/kozhurkin/async/pipers"
)

// tests: ✅
// bench: ⚠️ too slow

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
			printDebug("++ call i=%v, a=%v", i, a)
			value, err := f(i, a)
			printDebug("^ defer i=%v, a=%v, err=%v", i, a, err)
			if err == nil {
				<-traffic
			}
			return value, err
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
