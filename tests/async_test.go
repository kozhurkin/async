package tests

import (
	"context"
	"errors"
	"github.com/kozhurkin/async"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	ts := time.Now()
	pa := async.Pipeline(func() int { <-time.After(1 * time.Second); return 1 })
	pb := async.Pipeline(func() int { <-time.After(2 * time.Second); return 2 })
	a, b := <-pa, <-pb
	delta := int(time.Now().Sub(ts).Seconds())
	if delta != 2 {
		t.Fatal("Should complete in 2 seconds")
	}
	if a != 1 || b != 2 {
		t.Fatal("Wrong return values")
	}
	return
}

var throw = errors.New("throw error")
var throw2 = errors.New("throw error (2)")
var tasks = Tasks{
	{
		Desc: "Example of success launch        ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50 * time.Millisecond, nil},
			{100 * time.Millisecond, nil},
			{30 * time.Millisecond, nil},
			{40 * time.Millisecond, nil},
			{25 * time.Millisecond, nil},
		},
		Concurrency:    2,
		CancelAfter:    0,
		ExpectedResult: Result{1, 4, 9, 16, 25},
		ExpectedError:  nil,
	},
	{
		Desc: "Cancel context before throw      ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50 * time.Millisecond, nil},
			{100 * time.Millisecond, throw},
			{30 * time.Millisecond, nil},
			{40 * time.Millisecond, nil},
			{25 * time.Millisecond, nil},
		},
		Concurrency:    2,
		CancelAfter:    90 * time.Millisecond,
		ExpectedResult: Result{1, 0, 9, 0, 0},
		ExpectedError:  context.DeadlineExceeded,
	},
	{
		Desc: "Throw 1 error simple             ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50 * time.Millisecond, nil},
			{100 * time.Millisecond, throw},
			{30 * time.Millisecond, nil},
			{40 * time.Millisecond, nil},
			{25 * time.Millisecond, nil},
		},
		Concurrency:    2,
		CancelAfter:    0,
		ExpectedResult: Result{1, 0, 9, 0, 0},
		ExpectedError:  throw,
	},
	{
		Desc: "Throw 1 error before canceling   ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50 * time.Millisecond, nil},
			{100 * time.Millisecond, throw},
			{30 * time.Millisecond, nil},
			{40 * time.Millisecond, nil},
			{25 * time.Millisecond, nil},
		},
		Concurrency:    2,
		CancelAfter:    110 * time.Millisecond,
		ExpectedResult: Result{1, 0, 9, 0, 0},
		ExpectedError:  throw,
	},
	{
		Desc: "Throw 2 errors after each other  ",
		Args: [5]int{1, 2, 3, 4, 5},
		ProcessInfo: ProcessInfo{
			{50 * time.Millisecond, nil},
			{100 * time.Millisecond, throw},
			{30 * time.Millisecond, throw2},
			{40 * time.Millisecond, nil},
			{25 * time.Millisecond, nil},
		},
		Concurrency:    2,
		CancelAfter:    0,
		ExpectedResult: Result{1, 0, 0, 0, 0},
		ExpectedError:  throw2,
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
