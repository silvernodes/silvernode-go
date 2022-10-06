package stlutil

import (
	"sync"
)

type Map[K, V comparable] struct {
	_map     map[K]V
	breaking bool
	sync.RWMutex
}

func NewMap[K, V comparable]() *Map[K, V] {
	dict := Map[K, V]{}
	dict._map = make(map[K]V)
	return &dict
}

func NewMapRaw[K, V comparable](raw map[K]V) *Map[K, V] {
	dict := Map[K, V]{}
	dict._map = raw
	return &dict
}

func (d *Map[K, V]) Add(key K, value V) bool {
	d.Lock()
	defer d.Unlock()

	_, exist := d._map[key]
	if exist {
		return false
	}
	d._map[key] = value
	return true
}

func (d *Map[K, V]) Remove(key K) bool {
	d.Lock()
	defer d.Unlock()

	_, exist := d._map[key]
	if exist {
		delete(d._map, key)
		return true
	}
	return false
}

func (d *Map[K, V]) Set(key K, value V) {
	d.Lock()
	defer d.Unlock()

	d._map[key] = value
}

func (d *Map[K, V]) Get(key K) (V, bool) {
	d.RLock()
	defer d.RUnlock()

	v, exist := d._map[key]
	return v, exist
}

func (d *Map[K, V]) Len() int {
	d.RLock()
	defer d.RUnlock()

	return len(d._map)
}

func (d *Map[K, V]) ContainsKey(key K) bool {
	d.RLock()
	defer d.RUnlock()

	_, exist := d._map[key]
	return exist
}

func (d *Map[K, V]) ContainsValue(value V) bool {
	d.RLock()
	defer d.RUnlock()

	for _, v := range d._map {
		if v == value {
			return true
		}
	}
	return false
}

func (d *Map[K, V]) ForEach(fun func(K, V)) {
	d.RLock()
	defer d.RUnlock()

	d.breaking = false
	for k, v := range d._map {
		if d.breaking {
			break
		}
		fun(k, v)
	}
}

func (d *Map[K, V]) Break() {
	d.breaking = true
}

func (d *Map[K, V]) KeyValuePairs() map[K]V {
	d.RLock()
	defer d.RUnlock()

	ret := make(map[K]V)
	for k, v := range d._map {
		ret[k] = v
	}
	return ret
}

func (d *Map[K, V]) Raw() map[K]V {
	return d._map
}

func (d *Map[K, V]) Clear() {
	d.Lock()
	defer d.Unlock()

	for key, _ := range d._map {
		delete(d._map, key)
	}
}
