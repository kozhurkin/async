package tests

import (
	"context"
	"errors"
	"github.com/kozhurkin/async"
	"testing"
	"time"
)

var throw = errors.New("throw error")
var throw2 = errors.New("throw error (2)")
var tasks = Tasks{
	{
		Desc: "Example of success launch        ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, nil},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 0,
		TimeUnit:    time.Millisecond,
		Expectations: Expectations{
			{1, 250, Result{1, 4, 9, 16, 25}, nil},
			{2, 125, Result{1, 4, 9, 16, 25}, nil},
			{6, 100, Result{1, 4, 9, 16, 25}, nil},
		},
	},
	{
		Desc: "Cancel context before throw      ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 90,
		TimeUnit:    time.Millisecond,
		Expectations: Expectations{
			{1, 90, Result{1, 0, 0, 0, 0}, context.DeadlineExceeded},
			{2, 90, Result{1, 0, 9, 0, 0}, context.DeadlineExceeded},
			{6, 90, Result{1, 0, 9, 16, 25}, context.DeadlineExceeded},
		},
	},
	{
		Desc: "Throw 1 error simple             ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 0,
		TimeUnit:    time.Millisecond,
		Expectations: Expectations{
			{1, 150, Result{1, 0, 0, 0, 0}, throw},
			{2, 100, Result{1, 0, 9, 0, 0}, throw},
			{6, 100, Result{1, 0, 9, 16, 25}, throw},
		},
	},
	{
		Desc: "Throw 1 error before canceling   ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 110,
		TimeUnit:    time.Millisecond,
		Expectations: Expectations{
			{1, 110, Result{1, 0, 0, 0, 0}, context.DeadlineExceeded},
			{2, 100, Result{1, 0, 9, 0, 0}, throw},
			{6, 100, Result{1, 0, 9, 16, 25}, throw},
		},
	},
	{
		Desc: "Throw 2 errors after each other  ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, throw2},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 0,
		TimeUnit:    time.Millisecond,
		Expectations: Expectations{
			{1, 150, Result{1, 0, 0, 0, 0}, throw},
			{2, 80, Result{1, 0, 0, 0, 0}, throw2},
			{6, 30, Result{0, 0, 0, 0, 25}, throw2},
		},
	},
}

func TestAsyncSemaphore(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncSemaphore[int, int]}.Run()
}

func TestAsyncWorkers(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncWorkers[int, int]}.Run()
}

func TestAsyncPipers(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncPipers[int, int]}.Run()
	<-time.After(time.Second)
}

func TestAsyncErrgroup(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncErrgroup[int, int]}.Run()
}

func TestAsyncPromiseCatch(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncPromiseCatch[int, int]}.Run()
}

func TestAsyncPromiseAtomic(t *testing.T) {
	Launcher{t, tasks, async.AsyncPromiseAtomic[int, int]}.Run()
}

func TestAsyncPromiseSync(t *testing.T) {
	Launcher{t, tasks, async.AsyncPromiseSync[int, int]}.Run()
}

func TestAsyncPromisePipes(t *testing.T) {
	Launcher{t, tasks, async.AsyncPromisePipes[int, int]}.Run()
}
