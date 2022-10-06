package channel

import (
	"sync"

	"github.com/silvernodes/silvernode-go/process"
)

type Channel interface {
	Name() string
	Publish(datas interface{})
	Subscribe(fun func(interface{}), processor process.Processor) int64
	Unsubscribe(seq int64)
	UnsubscribeAll()
	NumOfSubscriber() int
}

var _chans map[string]*msgchannel = nil
var _lock sync.RWMutex

func init() {
	_chans = make(map[string]*msgchannel)
}

func Chan(ch string) Channel {
	_lock.Lock()
	defer _lock.Unlock()

	old, ok := _chans[ch]
	if ok {
		return old
	}

	_new := cochannel(ch)
	_chans[ch] = _new
	return _new
}

func Get(ch string) (Channel, bool) {
	old, ok := _chans[ch]
	return old, ok
}

func Close(ch string) {
	_lock.Lock()
	defer _lock.Unlock()

	old, ok := _chans[ch]
	if ok {
		old.UnsubscribeAll()
		delete(_chans, ch)
	}
}
