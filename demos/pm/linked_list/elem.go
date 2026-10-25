package main

type Elem[T any] struct {
	Val  T
	Prev *Elem[T]
	Next *Elem[T]
}
