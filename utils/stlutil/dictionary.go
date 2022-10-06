package stlutil

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type Dictionary[K, V comparable] struct {
	_keys    []K
	_map     map[K]V
	breaking bool
	comparer func(one, another V) bool
	sync.RWMutex
}

func NewDictionary[K, V comparable]() *Dictionary[K, V] {
	dict := Dictionary[K, V]{}
	dict._keys = make([]K, 0, 0)
	dict._map = make(map[K]V)
	return &dict
}

func (d *Dictionary[K, V]) Add(key K, value V) bool {
	d.Lock()
	defer d.Unlock()

	_, exist := d._map[key]
	if exist {
		return false
	}
	d._keys = append(d._keys, key)
	d._map[key] = value
	return true
}

func (d *Dictionary[K, V]) Remove(key K) bool {
	d.Lock()
	defer d.Unlock()

	_, exist := d._map[key]
	if exist {
		delete(d._map, key)

		index := -1
		for i, k := range d._keys {
			if k == key {
				index = i
				break
			}
		}
		if index > 0 {
			d._keys = append(d._keys[:index], d._keys[index+1:]...)
		}
		return true
	}
	return false
}

func (d *Dictionary[K, V]) Set(key K, value V) {
	d.Lock()
	defer d.Unlock()

	_, exist := d._map[key]
	if !exist {
		d._keys = append(d._keys, key)
	}
	d._map[key] = value
}

func (d *Dictionary[K, V]) Get(key K) (V, bool) {
	d.RLock()
	defer d.RUnlock()

	v, exist := d._map[key]
	return v, exist
}

func (d *Dictionary[K, V]) Len() int {
	d.RLock()
	defer d.RUnlock()

	return len(d._map)
}

func (d *Dictionary[K, V]) ContainsKey(key K) bool {
	d.RLock()
	defer d.RUnlock()

	_, exist := d._map[key]
	return exist
}

func (d *Dictionary[K, V]) ContainsValue(value V) bool {
	d.RLock()
	defer d.RUnlock()

	for _, v := range d._map {
		if v == value {
			return true
		}
	}
	return false
}

func (d *Dictionary[K, V]) GetKey(index int) (K, error) {
	d.RLock()
	defer d.RUnlock()

	if index < 0 || index >= len(d._keys) {
		return Default[K](), errors.New("Index Out Of Range!")
	}

	return d._keys[index], nil
}

func (d *Dictionary[K, V]) GetValue(index int) (V, error) {
	d.RLock()
	defer d.RUnlock()

	if index < 0 || index >= len(d._keys) {
		return Default[V](), errors.New("Index Out Of Range!")
	}

	k := d._keys[index]
	v, ok := d._map[k]
	if !ok {
		return Default[V](), errors.New("Invalid Key Value!")
	}
	return v, nil
}

func (d *Dictionary[K, V]) ForEach(fun func(int, K, V)) {
	d.RLock()
	defer d.RUnlock()

	d.breaking = false
	for i, k := range d._keys {
		if d.breaking {
			break
		}
		v, _ := d._map[k]
		fun(i, k, v)
	}
}

func (d *Dictionary[K, V]) Break() {
	d.breaking = true
}

func (d *Dictionary[K, V]) Clear() {
	d.Lock()
	defer d.Unlock()

	for key, _ := range d._map {
		delete(d._map, key)
	}

	d._keys = d._keys[0:0]
}

func (d *Dictionary[K, V]) Swap(index, index2 int) {
	d.Lock()
	defer d.Unlock()

	if index > len(d._keys) || index2 > len(d._keys) {
		fmt.Println("Dictionary Swap Failed!! Index Out Of Range!")
		return
	}

	d._keys[index], d._keys[index2] = d._keys[index2], d._keys[index]
}

func (d *Dictionary[K, V]) Less(index, index2 int) bool {
	d.RLock()
	defer d.RUnlock()

	if d.comparer == nil {
		return false
	}

	one, err := d.GetValue(index)
	another, err2 := d.GetValue(index2)

	if err != nil || err2 != nil {
		fmt.Println("Dictionary Less Failed!! Index Out Of Range!")
		return false
	}

	return d.comparer(one, another)
}

func (d *Dictionary[K, V]) SetComparer(fun func(one, another V) bool) {
	d.comparer = fun
}

func (d *Dictionary[K, V]) Sort() {
	sort.Sort(d)
}
