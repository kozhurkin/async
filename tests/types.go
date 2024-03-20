package tests

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
)

type ProcessInfo [5]struct {
	Delay int
	Err   error
}

type Expectations []struct {
	Concurrency int
	Duration    int
	Result
	Error error
}

type Tasks []struct {
	Desc string
	Args [5]int
	ProcessInfo
	CancelAfter int
	TimeUnit    time.Duration
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
				ctx, cancel = context.WithTimeout(ctx, time.Duration(task.CancelAfter)*task.TimeUnit)
				defer cancel()
			}
			ts := time.Now()
			result, err := l.Handler(ctx, task.Args[:5], func(i int, arg int) (int, error) {
				pi := task.ProcessInfo[i]
				<-time.After(time.Duration(pi.Delay) * task.TimeUnit)
				if pi.Err != nil {
					return 0, pi.Err
				}
				return arg * arg, nil
			}, expect.Concurrency)

			//l.T.Log(result, err)
			//assert.Equalf(l.T, task.ExpectedError, err, "__")
			//assert.Equal(l.T, task.ExpectedResult, result5)

			duration := int(time.Since(ts) / task.TimeUnit)
			diff := expect.Duration - duration

			l.T.Log(fmt.Sprintf(
				"%v :  c=%v \t %v %v \t %v %v      %v (%v)",
				task.Desc,
				expect.Concurrency,
				formatBool(math.Abs(float64(diff)) < 5),
				duration,
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
