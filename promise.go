package async

func Promise[R any, P struct {
	Start func() (chan R, chan error)
	Out   chan R
	Err   chan error
}](f func() (R, error)) P {
	out := make(chan R, 1)
	err := make(chan error, 1)

	res := P{Start: func() (chan R, chan error) {
		go func() {
			v, e := f()
			out <- v
			err <- e
			close(out)
			close(err)
		}()
		return out, err
	}, Out: out, Err: err}

	return res
}
