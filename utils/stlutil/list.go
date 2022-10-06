package stlutil

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type List[T comparable] struct {
	array    []T
	breaking bool
	comparer func(one, another T) bool
	sync.RWMutex
}

func NewList[T comparable](capacity int) *List[T] {
	list := List[T]{}
	list.array = make([]T, 0, capacity)
	return &list
}

func NewListRaw[T comparable](raw []T) *List[T] {
	list := List[T]{}
	list.array = raw
	return &list
}

func (l *List[T]) Add(item T) {
	l.Lock()
	defer l.Unlock()

	l.array = append(l.array, item)
}

func (l *List[T]) Insert(index int, item T) error {
	l.Lock()
	defer l.Unlock()

	if index > len(l.array) {
		return errors.New("Index Out Of Range!")
	}

	temp := make([]T, 0)
	after := append(temp, l.array[index:]...)
	before := l.array[0:index]
	l.array = append(before, item)
	l.array = append(l.array, after...)
	return nil
}

func (l *List[T]) RemoveAt(index int) error {
	l.Lock()
	defer l.Unlock()

	if index > len(l.array) {
		return errors.New("Index Out Of Range!")
	}

	l.array = append(l.array[:index], l.array[index+1:]...)
	return nil
}

func (l *List[T]) Remove(item T) bool {
	index := l.IndexOf(item)
	if index < 0 {
		return false
	}
	l.RemoveAt(index)
	return true
}

func (l *List[T]) IndexOf(item T) int {
	l.RLock()
	defer l.RUnlock()

	for index, one := range l.array {
		if one == item {
			return index
		}
	}
	return -1
}

func (l *List[T]) Contains(item T) bool {
	return l.IndexOf(item) >= 0
}

func (l *List[T]) Len() int {
	l.RLock()
	defer l.RUnlock()

	return len(l.array)
}

func (l *List[T]) Capacity() int {
	l.RLock()
	defer l.RUnlock()

	return cap(l.array)
}

func (l *List[T]) Items() []T {
	l.RLock()
	defer l.RUnlock()

	ret := make([]T, 0, len(l.array))
	for _, item := range l.array {
		ret = append(ret, item)
	}
	return ret
}

func (l *List[T]) Raw() []T {
	return l.array
}

func (l *List[T]) Get(index int) (T, error) {
	l.RLock()
	defer l.RUnlock()

	if index < 0 || index >= len(l.array) {
		return Default[T](), errors.New("Index Out Of Range!")
	}
	return l.array[index], nil
}

func (l *List[T]) Set(index int, item T) error {
	l.Lock()
	defer l.Unlock()

	if index > len(l.array) {
		return errors.New("Index Out Of Range!")
	}
	l.array[index] = item
	return nil
}

func (l *List[T]) ForEach(fun func(T)) {
	l.RLock()
	defer l.RUnlock()

	l.breaking = false
	for _, v := range l.array {
		if l.breaking {
			break
		}
		fun(v)
	}
}

func (l *List[T]) Break() {
	l.breaking = true
}

func (l *List[T]) Clear() {
	l.Lock()
	defer l.Unlock()

	l.array = l.array[0:0]
}

func (l *List[T]) Swap(index, index2 int) {
	l.Lock()
	defer l.Unlock()

	if index < 0 || index > len(l.array) || index2 < 0 || index2 > len(l.array) {
		fmt.Println("List Swap Failed!! Index Out Of Range!")
		return
	}

	l.array[index], l.array[index2] = l.array[index2], l.array[index]
}

func (l *List[T]) Less(index, index2 int) bool {
	l.RLock()
	defer l.RUnlock()

	if index < 0 || index > len(l.array) || index2 < 0 || index2 > len(l.array) {
		fmt.Println("List Less Failed!! Index Out Of Range!")
		return false
	}

	if l.comparer == nil {
		return false
	}

	return l.comparer(l.array[index], l.array[index2])

}

func (l *List[T]) SetComparer(fun func(one, another T) bool) {
	l.comparer = fun
}

func (l *List[T]) Sort() {
	sort.Sort(l)
}
