package async

type P[R any] struct {
	job    func() (R, error)
	Out    chan *R
	Err    chan error
	runned chan struct{}
	done   chan struct{}
	parent *P[R]
	//TODO error, result
}

func (p P[R]) markAsDone() {
	close(p.done)
}
func (p P[R]) markAsRunned() {
	close(p.runned)
}
func (p P[R]) Done() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}
func (p P[R]) Runned() bool {
	select {
	case <-p.runned:
		return true
	default:
		return false
	}
}

func (p *P[R]) Value() R {
	//fmt.Println("Value()", p)
	return *(<-p.Run().Out)
}
func (p *P[R]) Error() error {
	return <-p.Run().Err
}

func (p *P[R]) Return() (R, error) {
	err := p.Error()
	return p.Value(), err

}

func (p *P[R]) Then(f func(R) (R, error)) *P[R] {
	//fmt.Println("Then()", p)
	return p.join(f)
}

func (p *P[R]) join(f func(R) (R, error)) *P[R] {
	return Promise(func() (R, error) {
		err := p.Error()
		value := p.Value()
		if err != nil {
			return value, err
		}
		return f(value)
	})
}

func (p *P[R]) Run() *P[R] {
	//fmt.Println("Run", p)
	if p.parent != nil {
		go p.parent.Run()
	}
	if !p.Runned() {
		p.markAsRunned()
	} else {
		return p
	}
	go func() {
		v, e := p.job()
		p.Out <- &v
		p.Err <- e
		close(p.Out)
		close(p.Err)
		p.markAsDone()
	}()
	return p
}

func Resolve[R any](v R) *P[R] {
	return Promise(func() (R, error) {
		return v, nil
	}) //.Run()
}
func Promise[R any](job func() (R, error)) *P[R] {
	out := make(chan *R, 1)
	err := make(chan error, 1)

	res := P[R]{Out: out, Err: err}
	res.runned = make(chan struct{}, 1)
	res.done = make(chan struct{}, 1)
	res.job = job

	return &res
}
