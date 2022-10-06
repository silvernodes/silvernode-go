package pools

import (
	"sync"

	"github.com/silvernodes/silvernode-go/process"
)

var _pools map[string]*pool = nil
var _lock sync.RWMutex
var _svc process.Service = nil

func init() {
	_pools = make(map[string]*pool)
}

type Pool interface {
	Name() string
	Put(obj Disposable, life int64)
	NumOfObj() int
	Dispose(obj Disposable)
	DisposeAll()
}

type Disposable interface {
	Dispose()
}

func Fetch(name string) Pool {
	_lock.Lock()
	defer _lock.Unlock()

	if old, ok := _pools[name]; ok {
		return old
	}

	_new := copool(name)
	_pools[name] = _new

	if _svc == nil {
		_svc = process.SpawnS()
		_svc.StartTick(checkAll, 1000, nil)
	}

	return _new
}

func Get(name string) (Pool, bool) {
	_lock.Lock()
	defer _lock.Unlock()

	old, ok := _pools[name]
	return old, ok
}

func Dispose(name string) {
	_lock.Lock()
	defer _lock.Unlock()

	if old, ok := _pools[name]; ok {
		old.DisposeAll()
		delete(_pools, name)

		if len(_pools) <= 0 {
			_svc.Terminate()
			_svc = nil
		}
	}
}

func Summarize() map[string]int {
	_lock.RLock()
	defer _lock.RUnlock()

	sum := make(map[string]int)
	for name, po := range _pools {
		sum[name] = po.NumOfObj()
	}
	return sum
}

func checkAll() {
	_lock.RLock()
	defer _lock.RUnlock()

	for _, po := range _pools {
		po.checkAndDispose()
	}
}
