package pools

import (
	"sync"

	"github.com/silvernodes/silvernode-go/utils/errutil"
	"github.com/silvernodes/silvernode-go/utils/timeutil"
)

type pool struct {
	name string
	objs map[Disposable]int64
	sync.RWMutex
}

func copool(name string) *pool {
	p := new(pool)
	p.name = name
	p.objs = make(map[Disposable]int64)
	return p
}

func (p *pool) Name() string {
	return p.name
}

func (p *pool) Put(obj Disposable, life int64) {
	p.Lock()
	defer p.Unlock()

	p.objs[obj] = timeutil.Time() + life
}

func (p *pool) Dispose(obj Disposable) {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.objs[obj]; ok {
		obj.Dispose()
		delete(p.objs, obj)
	}
}

func (p *pool) NumOfObj() int {
	p.RLock()
	defer p.RUnlock()

	return len(p.objs)
}

func (p *pool) DisposeAll() {
	p.Lock()
	defer p.Unlock()

	for obj, _ := range p.objs {
		obj.Dispose()
	}
	p.objs = nil
	p.objs = make(map[Disposable]int64)
}

func (p *pool) checkAndDispose() {
	now := timeutil.Time()
	dirtys := make([]Disposable, 0, 10)
	p.RLock()
	for obj, ts := range p.objs {
		if now >= ts {
			dirtys = append(dirtys, obj)
		}
	}
	p.RUnlock()
	if len(dirtys) > 0 {
		p.Lock()
		for _, dirty := range dirtys {
			errutil.Try(dirty.Dispose, nil)
			delete(p.objs, dirty)
		}
		p.Unlock()
	}
}
