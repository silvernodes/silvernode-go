package process

import (
	"github.com/silvernodes/silvernode-go/utils/errutil"
)

type executor struct {
	// goroutine
	tasks      chan func()
	capacity   int
	terminated bool
}

func coexecutor(capacity int) *executor {
	e := new(executor)
	e.capacity = capacity
	e.tasks = make(chan func(), e.capacity)
	return e
}

func (e *executor) boot() {
	e.terminated = false
	defer func() {
		close(e.tasks)
		e.tasks = nil

	}()
	for task := range e.tasks {
		if e.terminated && task == nil {
			break
		}
		if task != nil {
			errutil.Try(task, nil)
		}
	}
}

func (e *executor) tasklen() int {
	if e.tasks != nil {
		return len(e.tasks)
	}
	return 0
}

func (e *executor) terminate() {
	e.terminated = true
	if e.tasks != nil {
		e.tasks <- nil
	}
}
