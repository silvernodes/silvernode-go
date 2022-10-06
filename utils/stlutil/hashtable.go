package stlutil

import (
	"sync"
)

type HashTable[K comparable, V any] struct {
	_map     map[K]V
	breaking bool
	sync.RWMutex
}

func NewHashTable[K comparable, V any]() *HashTable[K,V] {
	dict := HashTable[K,V]{}
	dict._map = make(map[K]V)
	return &dict
}

func NewHashTableRaw[K comparable, V any](raw map[K]V) *HashTable[K,V] {
	dict := HashTable[K,V]{}
	dict._map = raw
	return &dict
}

func (h *HashTable[K,V]) Add(key K, value V) bool {
	h.Lock()
	defer h.Unlock()

	_, exist := h._map[key]
	if exist {
		return false
	}
	h._map[key] = value
	return true
}

func (h *HashTable[K,V]) Remove(key K) bool {
	h.Lock()
	defer h.Unlock()

	_, exist := h._map[key]
	if exist {
		delete(h._map, key)
		return true
	}
	return false
}

func (h *HashTable[K,V]) Set(key K, value V) {
	h.Lock()
	defer h.Unlock()

	h._map[key] = value
}

func (h *HashTable[K,V]) Get(key K) (V, bool) {
	h.RLock()
	defer h.RUnlock()

	v, exist := h._map[key]
	return v, exist
}

func (h *HashTable[K,V]) Len() int {
	h.RLock()
	defer h.RUnlock()

	return len(h._map)
}

func (h *HashTable[K,V]) ContainsKey(key K) bool {
	h.RLock()
	defer h.RUnlock()

	_, exist := h._map[key]
	return exist
}

func (h *HashTable[K,V]) ForEach(fun func(K, V)) {
	h.RLock()
	defer h.RUnlock()

	h.breaking = false
	for k, v := range h._map {
		if h.breaking {
			break
		}
		fun(k, v)
	}
}

func (h *HashTable[K,V]) Break() {
	h.breaking = true
}

func (h *HashTable[K,V]) KeyValuePairs() map[K]V {
	h.RLock()
	defer h.RUnlock()

	ret := make(map[K]V)
	for k, v := range h._map {
		ret[k] = v
	}
	return ret
}

func (h *HashTable[K,V]) Raw() map[K]V {
	return h._map
}

func (h *HashTable[K,V]) Clear() {
	h.Lock()
	defer h.Unlock()

	for key, _ := range h._map {
		delete(h._map, key)
	}
}
