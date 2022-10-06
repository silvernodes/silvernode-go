package process

import (
	"sync"
)

type processor struct {
	pid     int64
	exes    []*executor
	killing bool
	multi   int
	sync.RWMutex
}

func coprocessor(pid int64, capacity int, multi int) *processor {
	p := new(processor)
	p.pid = pid
	if multi > 0 {
		p.exes = make([]*executor, 0, multi)
		p.killing = false
		p.multi = multi
		for i := 0; i < multi; i++ {
			exe := coexecutor(capacity)
			go exe.boot()
			p.exes = append(p.exes, exe)
		}
	}
	return p
}

func (p *processor) Pid() int64 {
	return p.pid
}

func (p *processor) Execute(task func()) bool {
	p.RLock()
	defer p.RUnlock()

	if p.killing {
		return false
	}
	if p.multi <= 0 {
		if task != nil {
			task()
		}
		return true
	}
	if p.multi == 1 {
		p.exes[0].tasks <- task
		return true
	}
	index := 0
	min := 999999
	for i, exe := range p.exes {
		num := exe.tasklen()
		if num < min {
			min = num
			index = i
		}
	}
	p.exes[index].tasks <- task
	return true
}

func (p *processor) Terminate() {
	p.RLock()
	defer p.RUnlock()

	if !p.killing {
		p.killing = true
		for _, exe := range p.exes {
			exe.terminate()
		}
		remove(p.pid)
	}
}

func (p *processor) Running() bool {
	return !p.killing
}

func (p *processor) TaskLen() int {
	p.RLock()
	defer p.RUnlock()

	sum := 0
	for _, exe := range p.exes {
		sum += exe.tasklen()
	}
	return sum
}

func (p *processor) CoroutineNum() int {
	return p.multi
}
