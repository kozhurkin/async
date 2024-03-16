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
	Desc string
	Args [5]int
	ProcessInfo
	Concurrency    int
	CancelAfter    time.Duration
	ExpectedResult Result
	ExpectedError  error
}

type Launcher struct {
	T *testing.T
	Tasks
	Handler func(context.Context, []int, func(int, int) (int, error), int) ([]int, error)
}

func (l Launcher) Run() *Launcher {
	for _, task := range l.Tasks {
		task := task
		//for _, c := range []int{task.Concurrency} {
		for _, c := range []int{1, task.Concurrency, len(task.Args) + 1} {
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
			}, c)

			//l.T.Log(result, err)
			//assert.Equalf(l.T, task.ExpectedError, err, "__")
			//assert.Equal(l.T, task.ExpectedResult, result5)

			l.T.Log(fmt.Sprintf(
				"%v :  c=%v \t %v \t %v %v, \t\t %v (%v)",
				task.Desc,
				c,
				time.Since(ts).Milliseconds(),
				task.ExpectedResult.IsEqual(result),
				result,
				errors.Is(err, task.ExpectedError),
				err,
			))
		}
		l.T.Log()
	}
	return &l
}

type Result [5]int

func (r Result) IsEqual(m []int) bool {
	//fmt.Println(r.String(), Result(m).String())
	if len(r) != len(m) {
		return false
	}
	for i, _ := range r {
		if r[i] != m[i] {
			return false
		}
	}
	return true
}
