package async

import (
	"fmt"
)

func Pip[V any, C chan V](f func() V) C {
	out := make(C, 1)
	go func() {
		out <- f()
		close(out)
	}()
	return out
}

func PipsReducer[V any](pips ...chan V) []V {
	res := make([]V, len(pips))
	for i, p := range pips {
		fmt.Printf("%v %v %T\n", i, p, p)
		res[i] = <-p
	}
	return res
}
