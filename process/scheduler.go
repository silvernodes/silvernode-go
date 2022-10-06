package process

import (
	"time"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

type scheduler struct {
	terminated bool
}

func coscheduler() *scheduler {
	s := new(scheduler)
	return s
}

func (s *scheduler) schedule(task func(), interval int, repeat int, delay int, callback func()) {
	s.terminated = false
	num := 0
	if delay > 0 {
		<-time.After(time.Millisecond * time.Duration(delay))
	}
	for {
		if s.terminated {
			break
		}
		errutil.Try(task, nil)
		if repeat > 0 {
			num++
			if num >= repeat {
				if callback != nil {
					callback()
				}
				s.terminate()
			}
		}
		if interval > 0 {
			<-time.After(time.Millisecond * time.Duration(interval))
		}
	}
}

func (s *scheduler) terminate() {
	s.terminated = true
}

func (s *scheduler) Cancel() {
	s.terminate()
}
