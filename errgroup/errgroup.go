package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	t := time.Now()
	data := []int{1, 2, 3, 4, 5}
	for i := 1; i <= 10000; i++ {
		res, err := async3(data)
		fmt.Println(res, err, time.Now().Sub(t))
	}
	<-time.After(time.Second)
}

func async2(data []int) ([]int, error) {
	ch := make(chan [2]int)
	defer close(ch)

	g := errgroup.Group{}
	for i, value := range data {
		i := i
		value := value
		g.Go(func() (err error) {
			fmt.Println("g.Go()", i, value)
			if value == rand.Intn(10) {
				err = errors.New("bad error")
				fmt.Println("THROW", err)
			}
			<-time.After(time.Duration(value) * time.Millisecond)
			ch <- [2]int{i, value}
			return err
		})
	}

	res := make([]int, len(data))
	go func() {
		for m := range ch {
			res[m[0]] = m[1]
		}
	}()

	if err := g.Wait(); err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}

	return res, nil
}

func async(data []int) ([]int, error) {
	res := make([]int, len(data))
	ch := make(chan struct {
		Key   int
		Value int
		error
	}, len(data))
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(2)
	for i, value := range data {
		i := i
		value := value
		g.Go(func() (err error) {
			fmt.Println("g.Go()", i, value)
			<-time.After(time.Duration(value*10) * time.Millisecond)
			if value == rand.Intn(10) {
				err = errors.New("bad error")
				fmt.Println("THROW", err)
			}
			select {
			case <-ctx.Done():
				fmt.Println("skipping", i, value)
			default:
				ch <- struct {
					Key   int
					Value int
					error
				}{i, value, err}
			}

			return err
		})
	}

	err := g.Wait()
	close(ch)

	if err != nil {
		fmt.Println("ERROR", err)
		fmt.Println("ERROR", ctx.Err())
		return nil, err
	}

	for m := range ch {
		res[m.Key] = m.Value
	}

	return res, nil
}

func async3(data []int) ([]int, error) {
	res := make([]int, len(data))
	ch := make(chan struct {
		Key   int
		Value int
		error
	}, len(data))
	//errchan := make(chan error)
	//g := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	for i, value := range data {
		i := i
		value := value
		go func() {
			var err error
			rnd := rand.Intn(10)
			fmt.Println("g.Go()", i, value)
			<-time.After(time.Duration(rnd*10) * time.Millisecond)
			if rand.Intn(len(data)) == 0 {
				err = errors.New("bad error")
				fmt.Println("THROW", err)
			}
			select {
			case <-ctx.Done():
				fmt.Println("skipping", i, value)
			default:
				ch <- struct {
					Key   int
					Value int
					error
				}{i, value, err}
			}
		}()
	}

	for range data {
		m := <-ch
		if m.error != nil {
			cancel()
			go func() {
				for v := range ch {
					fmt.Println(fmt.Sprintf("AAAAAAAAAAAAAAA %v %v %v %v", v.Key, v.Value, v.error, len(ch)))
					//<-time.After(time.Second)
				}
			}()
			return nil, m.error
		}
		res[m.Key] = m.Value
	}

	cancel()
	//close(ch)

	return res, nil
}
