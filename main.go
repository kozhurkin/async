package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

func main() {
	//fmt.Println(time.Now())
	//pa, pb := Pipeline(func() time.Time { return <-time.After(2 * time.Second) }),
	//	Pipeline(func() time.Time { return <-time.After(3 * time.Second) })
	//a, b := <-pa, <-pb
	//fmt.Println(time.Now(), []time.Time{a, b})
	//return
	rand.Seed(time.Now().UnixNano())

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
		res, err := MapPromise2(ctx, data, func(k int) (int, error) {
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
