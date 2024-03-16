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

type Expectations []struct {
	Concurrency int
	Result
	Error error
	//time.Duration
}

type Tasks []struct {
	Desc string
	Args [5]int
	ProcessInfo
	CancelAfter time.Duration
	Expectations
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
		for _, expect := range task.Expectations {
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
			}, expect.Concurrency)

			//l.T.Log(result, err)
			//assert.Equalf(l.T, task.ExpectedError, err, "__")
			//assert.Equal(l.T, task.ExpectedResult, result5)

			l.T.Log(fmt.Sprintf(
				"%v :  c=%v \t %v \t %v %v, \t %v (%v)",
				task.Desc,
				expect.Concurrency,
				time.Since(ts).Milliseconds(),
				formatBool(expect.Result.IsEqual(result)),
				result,
				formatBool(errors.Is(err, expect.Error)),
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

func formatBool(b bool) string {
	if b {
		//return "✔"
		return "✅"
	} else {
		//return "✖"
		return "❌"
	}
}
