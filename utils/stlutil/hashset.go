package stlutil

import (
	"sync"
)

type HashSet [T comparable] struct {
	_map     map[T]struct{}
	breaking bool
	sync.RWMutex
}

func NewHashSet[T comparable]() *HashSet[T] {
	h := HashSet[T]{}
	h._map = make(map[T]struct{})
	return &h
}

func NewHashSetRaw[T comparable](raw []T) *HashSet[T] {
	h := HashSet[T]{}
	h._map = make(map[T]struct{})
	for _, item := range raw {
		h._map[item] = struct{}{}
	}

	return &h
}

func (h *HashSet[T]) Add(item T) bool {
	h.Lock()
	defer h.Unlock()

	_, exist := h._map[item]
	if exist {
		return false
	}
	h._map[item] = struct{}{}
	return true
}

func (h *HashSet[T]) Remove(item T) bool {
	h.Lock()
	defer h.Unlock()

	_, exist := h._map[item]
	if exist {
		delete(h._map, item)
		return true
	}
	return false
}

func (h *HashSet[T]) Len() int {
	h.RLock()
	defer h.RUnlock()

	return len(h._map)
}

func (h *HashSet[T]) Contains(item T) bool {
	h.RLock()
	defer h.RUnlock()

	_, exist := h._map[item]
	return exist
}

func (h *HashSet[T]) ForEach(fun func(T)) {
	h.RLock()
	defer h.RUnlock()

	h.breaking = false
	for k, _ := range h._map {
		if h.breaking {
			break
		}
		fun(k)
	}
}

func (h *HashSet[T]) Break() {
	h.breaking = true
}

func (h *HashSet[T]) Items() []T {
	h.RLock()
	defer h.RUnlock()

	items := make([]T, 0, len(h._map))
	for k, _ := range h._map {
		items = append(items, k)
	}
	return items
}

func (h *HashSet[T]) Clear() {
	h.Lock()
	defer h.Unlock()

	for k, _ := range h._map {
		delete(h._map, k)
	}
}

func (h *HashSet[T]) Intersect(another *HashSet[T]) *HashSet[T] {
	clone := h.Clone()
	clone.IntersectWith(another)
	return clone
}

func (h *HashSet[T]) IntersectWith(another *HashSet[T]) {
	h.Lock()
	defer h.Unlock()

	for k, _ := range h._map {
		if !another.Contains(k) {
			delete(h._map, k)
		}
	}
}

func (h *HashSet[T]) Except(another *HashSet[T]) *HashSet[T] {
	clone := h.Clone()
	clone.ExceptWith(another)
	return clone
}

func (h *HashSet[T]) ExceptWith(another *HashSet[T]) {
	h.Lock()
	defer h.Unlock()

	for k, _ := range h._map {
		if another.Contains(k) {
			delete(h._map, k)
		}
	}
}

func (h *HashSet[T]) Union(another *HashSet[T]) *HashSet[T] {
	clone := h.Clone()
	clone.UnionWith(another)
	return clone
}

func (h *HashSet[T]) UnionWith(another *HashSet[T]) {
	another.ForEach(func(item T) {
		if !h.Contains(item) {
			h.Add(item)
		}
	})
}

func (h *HashSet[T]) IsSubset(another *HashSet[T]) bool {
	if h.Len() > another.Len() {
		return false
	}

	h.RLock()
	defer h.RUnlock()

	for k, _ := range h._map {
		if !another.Contains(k) {
			return false
		}
	}
	return true
}

func (h *HashSet[T]) IsSuperset(another *HashSet[T]) bool {
	return another.IsSubset(h)
}

func (h *HashSet[T]) Clone() *HashSet[T] {
	return NewHashSetRaw(h.Items())
}
