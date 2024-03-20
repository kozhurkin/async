package tests

import (
	"fmt"
	"github.com/kozhurkin/async/pip"
	"testing"
	"time"
)

func TestPip(t *testing.T) {
	ts := time.Now()
	pa := pip.NewPip(func() int { <-time.After(1 * time.Second); return 1 })
	pb := pip.NewPip(func() int { <-time.After(2 * time.Second); return 2 })
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
func TestPipsReducer(t *testing.T) {
	ts := time.Now()
	res := pip.PipsReducer(
		pip.NewPip(func() int { <-time.After(1 * time.Second); return 1 }),
		pip.NewPip(func() int { <-time.After(2 * time.Second); return 2 }),
	)
	fmt.Println(res, time.Since(ts))
	return
}
