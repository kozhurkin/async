package tests

import (
	"errors"
	"fmt"
	"github.com/kozhurkin/async/promise"
	"testing"
	"time"
)

func TestPromise(t *testing.T) {
	//for i := 0; i < 1000; i += 1 {
	ts := time.Now()
	p := promise.Resolve(2).Then(func(v int) (int, error) {
		<-time.After(10 * time.Millisecond)
		return v * v, errors.New("oops")
	}).Then(func(v int) (int, error) {
		<-time.After(20 * time.Millisecond)
		return v * v, nil //errors.New("oops")
	}).Then(func(v int) (int, error) {
		<-time.After(20 * time.Millisecond)
		fmt.Println("LAST")
		return v * v, nil
	}).Catch(func(e error) (int, error) {
		fmt.Println("CATCH", e)
		return 5, nil
	})

	//p.Run()
	//go p.Run()
	//p.Error()
	go func() {
		fmt.Println("goooooo", p.Value())
		<-time.After(time.Second)
		fmt.Println("goooooo", p.Value())
	}()

	fmt.Println(p.Return())
	//if p.Value() != 256 {
	//	panic("sadadas")
	//}
	fmt.Println()
	fmt.Println(p.Runned(), p.Completed(), time.Since(ts))
	<-time.After(time.Second)
	//}
}
