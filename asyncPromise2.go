package async

import "context"

// array of channels
func AsyncPromise2[A any, V any](ctx context.Context, args []A, f func(A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}
	printDebug("CONCURRENCY: %v", concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	promises := make([]chan V, len(args))

	traffic := make(chan struct{}, concurrency-1)
	end := make(chan error, concurrency)

LOOP:
	for i, arg := range args {
		i, arg := i, arg
		select {
		case <-ctx.Done():
			promises = promises[0:i]
			printDebug("SKIP %v", len(promises))
			break LOOP
		default:
		}
		promises[i] = Pipeline(func() V {
			printDebug("JOB START: i=%v arg=%v", i, arg)
			value, err := f(arg)
			if err != nil {
				end <- err
				cancel()
			}
			printDebug("JOB DONE: i=%v arg=%v value=%v err=%v", i, arg, value, err)
			<-traffic
			return value
		})
		printDebug("promises[%v] = p", i)
		traffic <- struct{}{}
	}
	printDebug("LOOP END")
	printDebug("close(traffic) %v", len(traffic))
	close(traffic)

	res := make([]V, len(args))
	for i, p := range promises {
		msg := <-p
		printDebug("fill %v %v %v", i, p, msg)
		res[i] = msg
	}

	close(end)
	err := <-end
	printDebug("err := %v", err)
	if err != nil {
		return res, err
	}

	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		return res, nil
	}
}
