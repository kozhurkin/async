package pip

func NewPip[V any, C chan V](f func() V) C {
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
		res[i] = <-p
	}
	return res
}
