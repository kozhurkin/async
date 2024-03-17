// You can edit this code!
// Click here and start typing.
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx2, cancel2 := context.WithCancel(ctx)
	defer cancel2()

	i := 0

	exit := time.After(time.Second)

OUT:
	for {

		select {
		case <-ctx.Done():
			fmt.Println("ctx")
		case <-ctx2.Done():
			fmt.Println("ctx2")
		case <-exit:
			fmt.Println("done")
			break OUT
		default:
			i += 1
			//fmt.Println(time.Now())

		}
	}
	fmt.Println("hi", i)
}
