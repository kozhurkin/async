package tests

import (
	"fmt"
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
func TestPipelineReducer(t *testing.T) {
	ts := time.Now()
	res := async.PipelineReducer(
		async.Pipeline(func() int { <-time.After(1 * time.Second); return 1 }),
		async.Pipeline(func() int { <-time.After(2 * time.Second); return 2 }),
	)
	fmt.Println(res, time.Since(ts))
	return
}
