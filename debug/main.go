package main

import (
	"context"
	"fmt"
	"github.com/504dev/async"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

func main() {
	fmt.Println(map[string]int{"aaa": 11e9, "bbb": 5.7e9})
	//fmt.Println(time.Now())
	//pa := async.Pipeline(func() time.Time { return <-time.After(2 * time.Second) })
	//pb := async.Pipeline(func() time.Time { return <-time.After(3 * time.Second) })
	//a, b := <-pa, <-pb
	//fmt.Println(time.Now(), []time.Time{a, b})
	//a2, b2 := <-pa, <-pb
	//fmt.Println(time.Now(), []time.Time{a2, b2})
	//return
	rand.Seed(time.Now().UnixNano())
	async.SetDebug(1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// handle SIGINT (control+c)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		select {
		case <-c:
			fmt.Println("\nmain: interrupt received. cancelling context.")
			cancel()
		}
	}()

	concurrency := 3
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
OUT:
	for i := 1; i <= 1; i++ {
		res, err := async.AsyncPromise(ctx, data, func(k int) (int, error) {
			rnd := rand.Intn(1000)
			<-time.After(time.Duration(rnd) * time.Millisecond)
			if rand.Intn(len(data)) == 0 {
				return k, fmt.Errorf("unknown error (%v)", k)
			}
			return k, nil
		}, concurrency)
		fmt.Printf("[%v] RESULT: %v %v\n", i, res, err)
		select {
		case <-ctx.Done():
			fmt.Println("BREAAAAAK")
			break OUT
		default:
			continue
		}
	}
	<-time.After(time.Duration(2) * time.Second)
}
