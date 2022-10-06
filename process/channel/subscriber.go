package channel

import (
	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/errutil"
)

type subscriber struct {
	seq       int64
	fun       func(interface{})
	processor process.Processor
}

func cosubscriber(seq int64, fun func(interface{}), processor process.Processor) *subscriber {
	s := new(subscriber)
	s.seq = seq
	s.fun = fun
	s.processor = processor
	return s
}

func (s *subscriber) publish(datas interface{}) {
	if s.processor != nil && s.processor.Running() {
		s.processor.Execute(func() {
			errutil.Try(func() {
				s.fun(datas)
			}, nil)
		})
	} else {
		errutil.Try(func() {
			s.fun(datas)
		}, nil)
	}
}
