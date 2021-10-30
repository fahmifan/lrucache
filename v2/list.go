package lrucache

import (
	"container/list"
	"sync"
)

// List is concurrent safe doubly linked list
type List struct {
	list  *list.List
	mutex *sync.RWMutex
}

func NewList() *List {
	return &List{
		list:  list.New(),
		mutex: &sync.RWMutex{},
	}
}

func (l *List) PushFront(item *Item) *Item {
	l.mutex.Lock()
	el := l.list.PushFront(item)
	item.SetElement(el)
	l.mutex.Unlock()
	return item
}

func (l *List) Front() *Item {
	l.mutex.RLock()
	el := l.list.Front()
	l.mutex.RUnlock()
	val, _ := el.Value.(*Item)
	return val
}

func (l *List) Back() *Item {
	l.mutex.RLock()
	el := l.list.Back()
	l.mutex.RUnlock()
	item, _ := el.Value.(*Item)
	return item
}

// special case
func (l *List) Remove(e *list.Element) *Item {
	l.mutex.Lock()
	val := l.list.Remove(e)
	l.mutex.Unlock()
	if val == nil {
		return nil
	}
	item, _ := val.(*Item)
	return item
}

// special case
func (l *List) RemoveBack() *Item {
	last := l.Back()
	item := l.Remove(last.element)
	return item
}

func (l *List) MoveToFront(item *Item) {
	l.mutex.Lock()
	l.list.MoveToFront(item.element)
	l.mutex.Unlock()
}
