package async

import (
	"fmt"
)

func Pipeline[R any, C chan R](f func() R) C {
	out := make(C, 1)
	go func() {
		out <- f()
		close(out)
	}()
	return out
}

func PipelineReducer[R any](pipes ...chan R) []R {
	res := make([]R, len(pipes))
	for i, p := range pipes {
		fmt.Printf("%v %v %T\n", i, p, p)
		res[i] = <-p
	}
	return res
}
