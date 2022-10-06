package process

import (
	"sync"
	"time"

	"github.com/silvernodes/silvernode-go/utils/snowflake"
)

var _processors map[int64]Process
var lock sync.RWMutex

func init() {
	_processors = make(map[int64]Process)
}

func Spawn(capacity int) Processor {
	lock.Lock()
	defer lock.Unlock()

	pid := snowflake.GenerateRaw()
	processor := coprocessor(pid, capacity, 1)
	_processors[pid] = processor
	return processor
}

func SpawnM(multi int, capacity int) Processor {
	lock.Lock()
	defer lock.Unlock()

	pid := snowflake.GenerateRaw()
	processor := coprocessor(pid, capacity, multi)
	_processors[pid] = processor
	return processor
}

func SpawnS() Service {
	lock.Lock()
	defer lock.Unlock()

	pid := snowflake.GenerateRaw()
	service := coservice(pid)
	_processors[pid] = service
	return service
}

func get(pid int64) (Process, bool) {
	lock.RLock()
	defer lock.RUnlock()

	exe, exists := _processors[pid]
	return exe, exists
}

func GetProcessor(pid int64) (Processor, bool) {
	item, exists := get(pid)
	if !exists {
		return nil, false
	}
	p, ok := item.(*processor)
	return p, ok
}

func GetService(pid int64) (Service, bool) {
	item, exists := get(pid)
	if !exists {
		return nil, false
	}
	s, ok := item.(*service)
	return s, ok
}

func Kill(pid int64) {
	if processor, ok := get(pid); ok {
		processor.Terminate()
	}
}

func Schedule(task func(), interval int, repeat int, delay int, callback func()) Scheduler {
	s := coscheduler()
	go s.schedule(task, interval, repeat, delay, callback)
	return s
}

func Go(task func() interface{}) Coroutine {
	c := cocoroutine()
	go c.do(task)
	return c
}

func remove(pid int64) {
	lock.Lock()
	defer lock.Unlock()

	delete(_processors, pid)
}

func ProcessNum() int {
	lock.RLock()
	defer lock.RUnlock()

	return len(_processors)
}

func CoroutineNum() int {
	lock.RLock()
	defer lock.RUnlock()

	sum := 0
	for _, item := range _processors {
		switch v := item.(type) {
		case *processor:
			sum += v.CoroutineNum()
		default:
			sum += 1
		}
	}
	return sum
}

func Sleep(ms int) {
	<-time.After(time.Millisecond * time.Duration(ms))
}
