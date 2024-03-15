package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type ProcessInfo [5]struct {
	Delay time.Duration
	Err   error
}

type Tasks []struct {
	Desc   string
	Args   [5]int
	Delays [5]time.Duration
	Errors [5]error
	ProcessInfo
	Concurrency    int
	CancelAfter    time.Duration
	ExpectedResult [5]int
	ExpectedError  error
}

type Launcher struct {
	T *testing.T
	Tasks
	Handler func(context.Context, []int, func(int, int) (int, error), int) ([]int, error)
}

func (l Launcher) Run() {
	for _, task := range l.Tasks {
		task := task
		ctx := context.Background()
		if task.CancelAfter != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, task.CancelAfter)
			defer cancel()
		}
		ts := time.Now()
		result, err := l.Handler(ctx, task.Args[:5], func(i int, arg int) (int, error) {
			pi := task.ProcessInfo[i]
			<-time.After(pi.Delay)
			if pi.Err != nil {
				return 0, pi.Err
			}
			return arg * arg, nil
		}, task.Concurrency)

		//l.T.Log(result, err)

		var result5 [5]int
		copy(result5[:], result)

		//assert.Equalf(l.T, task.ExpectedError, err, "__")
		//assert.Equal(l.T, task.ExpectedResult, result5)

		l.T.Log(fmt.Sprintf(
			"%v :  %v \t %v %v, \t\t %v (%v)",
			task.Desc,
			time.Since(ts).Milliseconds(),
			result5 == task.ExpectedResult,
			result,
			errors.Is(err, task.ExpectedError),
			err,
		))
	}
}
