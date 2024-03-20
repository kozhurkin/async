package main

import (
	"errors"
	"fmt"
	"github.com/kozhurkin/async/promise"
	"golang.org/x/sync/errgroup"
	"time"
)

func main() {
	ts := time.Now()
	fmt.Println("start")

	p2 := promise.NewPromise(func() (int, error) {
		<-time.After(2 * time.Second)
		return 2, nil
	})

	p1 := promise.NewPromise(func() (int, error) {
		<-time.After(1 * time.Second)
		return 1, nil
	})

	fmt.Println(p1.Runned(), p2.Runned())
	p1.Run().Run()
	fmt.Println(p1.Runned(), p2.Runned())
	p2.Run()
	fmt.Println(p1.Runned(), p2.Runned())

	fmt.Println(p1.Done(), p2.Done())

	r2, r3 := p2.Value(), p1.Value()

	fmt.Println("___", r2, r3, time.Since(ts).Seconds())
	ts = time.Now()

	fmt.Println("go")
	wg := new(errgroup.Group)

	wg.Go(func() (err error) {
		<-time.After(1 * time.Second)
		return errors.New("test")
	})
	wg.Go(func() (err error) {
		return errors.New("test2")
	})
	err := wg.Wait()
	fmt.Println(err, time.Since(ts).Seconds())
}
