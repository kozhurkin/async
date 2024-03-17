package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/kozhurkin/async"
	"testing"
	"time"
)

func TestPromise(t *testing.T) {
	p := async.Resolve(5).Then(func(v int) (int, error) {
		return v * v, nil //errors.New("oops")
	}).Then(func(v int) (int, error) {
		return v * v, nil
	})

	fmt.Println(p.Return())
}

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
func TestPipelineReducer(t *testing.T) {
	ts := time.Now()
	res := async.PipelineReducer(
		async.Pipeline(func() int { <-time.After(1 * time.Second); return 1 }),
		async.Pipeline(func() int { <-time.After(2 * time.Second); return 2 }),
	)
	fmt.Println(res, time.Since(ts))
	return
}

func TestPipers(t *testing.T) {
	ts := time.Now()
	pp := async.Pipers[int]{
		async.NewPiper(func() (int, error) { <-time.After(10 * time.Millisecond); return 1, nil }),
		async.NewPiper(func() (int, error) { <-time.After(20 * time.Millisecond); return 2, errors.New("surprise") }),
		async.NewPiper(func() (int, error) { <-time.After(15 * time.Millisecond); return 3, nil }),
		async.NewPiper(func() (int, error) { <-time.After(25 * time.Millisecond); return 4, errors.New("surprise 2") }),
		async.NewPiper(func() (int, error) { <-time.After(30 * time.Millisecond); return 5, nil }),
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 21*time.Millisecond)
	defer cancel()

	err := pp.ErrorsAllContext(ctx)

	res := pp.Results()

	fmt.Println(res, err, time.Since(ts))

	return
}
