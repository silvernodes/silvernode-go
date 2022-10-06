package process

type coroutine struct {
	signal chan interface{}
}

func cocoroutine() *coroutine {
	c := new(coroutine)
	c.signal = make(chan interface{}, 1)
	return c
}

func (c *coroutine) do(task func() interface{}) {
	c.signal <- task()
	close(c.signal)
}

func (c *coroutine) Sync() interface{} {
	return <-c.signal
}
