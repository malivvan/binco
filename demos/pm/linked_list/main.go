// Command linked_list encodes a doubly linked list; the pointers are restored, so the decoded list is fully functional.
package main

import (
	"fmt"

	"github.com/malivvan/binco/varnum"
)

func main() {
	list := ShortLinkedList()
	serializer := MakeLinkedListSerializer[int](varnum.Int)

	bs := make([]byte, serializer.Size(list))
	n := serializer.Marshal(list, bs)
	fmt.Printf("list:     %v (%d elements)\n", list, list.Len)
	fmt.Printf("size:     %d bytes\n", len(bs))
	fmt.Printf("encoded:  %v (wrote %d)\n", bs, n)

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded:  %v (read %d bytes)\n", back, n)
	fmt.Printf("forward:  %v\n", back.Forward())
	fmt.Printf("backward: %v\n", back.Backward())

	ok := true
	for e := back.Head; e != nil && e.Next != nil; e = e.Next {
		if e.Next.Prev != e {
			ok = false
		}
	}
	fmt.Printf("prev/next links consistent: %v\n", ok)
	fmt.Printf("Head.Prev == nil: %v, Tail.Next == nil: %v\n",
		back.Head.Prev == nil, back.Tail.Next == nil)

	second := back.Head.Next
	back.Remove(second)
	fmt.Printf("\nafter removing %v: %v (len=%d)\n", second.Val, back, back.Len)
	fmt.Printf("new head.Next.Prev == head: %v\n", back.Head.Next.Prev == back.Head)

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped:  %d bytes\n", n)
}

func ShortLinkedList() LinkedList[int] {
	var l LinkedList[int]
	for _, v := range []int{8, 9, 10, 11} {
		l.AddBack(v)
	}
	return l
}
