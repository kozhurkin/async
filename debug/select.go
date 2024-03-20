package main

import (
	"fmt"
	"time"
)

func main() {
	durations := []int{300, 100, 200}
	pipes := make([]chan int, len(durations))
	for i, d := range durations {
		i, d := i, d
		pipes[i] = make(chan int, 1)
		go func() {
			<-time.After(time.Duration(d) * time.Millisecond)
			pipes[i] <- d
		}()
	}

	result := make([]int, 0)
	i := 1
	defer func() {
		fmt.Println("result", result)
		fmt.Println("iterations", i)
	}()

	for {
		for _, p := range pipes {
			select {
			case v := <-p:
				fmt.Println(v)
				result = append(result, v)
				if len(result) == len(pipes) {
					return
				}
				//default: // comment or uncomment this line, see results
			}
			i += 1
		}
	}
}
