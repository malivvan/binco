package main

import (
	"strconv"
	"strings"
)

type LinkedList[T any] struct {
	Head *Elem[T]
	Tail *Elem[T]
	Len  int
}

func (l *LinkedList[T]) AddFront(v T) {
	elem := &Elem[T]{Val: v}
	if l.Head == nil {
		l.Tail = elem
	} else {
		elem.Next = l.Head
		l.Head.Prev = elem
	}
	l.Head = elem
	l.Len++
}

func (l *LinkedList[T]) AddBack(v T) {
	elem := &Elem[T]{Val: v}
	if l.Tail == nil {
		l.Head = elem
	} else {
		elem.Prev = l.Tail
		l.Tail.Next = elem
	}
	l.Tail = elem
	l.Len++
}

func (l *LinkedList[T]) Remove(e *Elem[T]) {
	switch {
	case e == l.Head && e == l.Tail:
		l.Head, l.Tail = nil, nil
	case e == l.Head:
		l.Head = e.Next
		l.Head.Prev = nil
	case e == l.Tail:
		l.Tail = e.Prev
		l.Tail.Next = nil
	default:
		e.Prev.Next = e.Next
		e.Next.Prev = e.Prev
	}
	l.Len--
}

func (l LinkedList[T]) Forward() (values []T) {
	for e := l.Head; e != nil; e = e.Next {
		values = append(values, e.Val)
	}
	return
}

func (l LinkedList[T]) Backward() (values []T) {
	for e := l.Tail; e != nil; e = e.Prev {
		values = append(values, e.Val)
	}
	return
}

func (l LinkedList[T]) String() string {
	parts := make([]string, 0, l.Len)
	for _, v := range l.Forward() {
		parts = append(parts, strconv.FormatInt(int64(toInt(v)), 10))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func toInt[T any](v T) int {
	if n, ok := any(v).(int); ok {
		return n
	}
	return 0
}
