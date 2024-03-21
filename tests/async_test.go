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
		Desc: "SUCCESS launch          ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, nil},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 0,
		TimeUnit:    2 * time.Millisecond,
		Expectations: Expectations{
			{1, 5, 245, Result{1, 4, 9, 16, 25}, nil},
			{2, 5, 125, Result{1, 4, 9, 16, 25}, nil},
			{6, 5, 100, Result{1, 4, 9, 16, 25}, nil},
		},
	},
	{
		Desc: "DEADLINE before THROW   ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 90,
		TimeUnit:    2 * time.Millisecond,
		Expectations: Expectations{
			{1, 2, 90, Result{1, 0, 0, 0, 0}, context.DeadlineExceeded},
			{2, 4, 90, Result{1, 0, 9, 0, 0}, context.DeadlineExceeded},
			{6, 5, 90, Result{1, 0, 9, 16, 25}, context.DeadlineExceeded},
		},
	},
	{
		Desc: "THROW 1 error simple    ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 0,
		TimeUnit:    2 * time.Millisecond,
		Expectations: Expectations{
			{1, 2, 150, Result{1, 0, 0, 0, 0}, throw},
			{2, 4, 100, Result{1, 0, 9, 0, 0}, throw},
			{6, 5, 100, Result{1, 0, 9, 16, 25}, throw},
		},
	},
	{
		Desc: "THROW 1 before DEADLINE ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, nil},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 110,
		TimeUnit:    2 * time.Millisecond,
		Expectations: Expectations{
			{1, 2, 110, Result{1, 0, 0, 0, 0}, context.DeadlineExceeded},
			{2, 4, 100, Result{1, 0, 9, 0, 0}, throw},
			{6, 5, 100, Result{1, 0, 9, 16, 25}, throw},
		},
	},
	{
		Desc: "THROW 2 errors following",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50, nil},
			{100, throw},
			{30, throw2},
			{40, nil},
			{25, nil},
		},
		CancelAfter: 0,
		TimeUnit:    2 * time.Millisecond,
		Expectations: Expectations{
			{1, 2, 150, Result{1, 0, 0, 0, 0}, throw},
			{2, 3, 80, Result{1, 0, 0, 0, 0}, throw2},
			{6, 5, 30, Result{0, 0, 0, 0, 25}, throw2},
		},
	},
}

func TestAsyncSemaphore(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncSemaphore[int, int]}.
		//Pick(4, 1).
		Run()
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

func TestAsyncPromiseAtomic(t *testing.T) {
	Launcher{t, tasks, async.AsyncPromiseAtomic[int, int]}.Run()
}

func TestAsyncErrgroup(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncErrgroup[int, int]}.Run()
}

func TestAsyncPromiseCatch(t *testing.T) {
	//async.SetDebug(1)
	Launcher{t, tasks, async.AsyncPromiseCatch[int, int]}.Run()
}

func TestAsyncPromiseSync(t *testing.T) {
	Launcher{t, tasks, async.AsyncPromiseSync[int, int]}.Run()
}

func TestAsyncPromisePipes(t *testing.T) {
	Launcher{t, tasks, async.AsyncPromisePipes[int, int]}.Run()
}
