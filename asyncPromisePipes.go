package async

import (
	"context"
	"github.com/kozhurkin/async/pip"
)

func AsyncPromisePipes[A any, V any](ctx context.Context, args []A, f func(context.Context, int, A) (V, error), concurrency int) ([]V, error) {
	if concurrency == 0 {
		concurrency = len(args)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	promises := make([]chan V, len(args))

	traffic := make(chan struct{}, concurrency)
	catch := make(chan error, 1)

LOOP:
	for i, arg := range args {
		i, arg := i, arg

		select {
		case <-ctx.Done():
			promises = promises[0:i]
			break LOOP
		case traffic <- struct{}{}:
		}

		promises[i] = pip.NewPip(func() V {
			defer func() { <-traffic }()
			value, err := f(ctx, i, arg)
			if err != nil {
				select {
				case catch <- err:
				default:
				}
				cancel()
			}
			return value
		})
	}

	res := make([]V, len(args))
	for i, p := range promises {
		select {
		case res[i] = <-p:
		case <-ctx.Done():
			select {
			case res[i] = <-p:
			default:
			}
		}
	}

	select {
	case err := <-catch:
		return res, err
	default:
	}

	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		return res, nil
	}
}
