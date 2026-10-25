package binco

import "unsafe"

// PtrFlag is the one-byte tag that introduces a pointer.
type PtrFlag byte

// The pointer flags: PtrNotNil is followed by the value, PtrNil ends the pointer, and PtrMapping is followed by a pointer identifier.
const (
	// PtrNotNil means that the pointer is not nil and its value follows.
	PtrNotNil PtrFlag = iota
	// PtrNil means that the pointer is nil.
	PtrNil
	// PtrMapping means that the pointer is encoded by reference: a positive
	// integer pointer ID follows, and the pointed-to data is present only the
	// first time the ID occurs.
	PtrMapping
)

// NewPtrMap returns an empty map from pointer to identifier.
func NewPtrMap() *PtrMap {
	return &PtrMap{
		ids: make(map[unsafe.Pointer]int),
	}
}

// PtrMap hands out increasing identifiers for pointers while a value is encoded.
type PtrMap struct {
	next int
	ids  map[unsafe.Pointer]int
}

// Put stores ptr and returns its identifier.
func (m *PtrMap) Put(ptr unsafe.Pointer) (key int) {
	key = m.next
	m.ids[ptr] = key
	m.next++
	return
}

// Get reports the identifier of ptr, if it has one.
func (m *PtrMap) Get(ptr unsafe.Pointer) (key int, pst bool) {
	key, pst = m.ids[ptr]
	return
}

// Len reports how many pointers have been stored.
func (m *PtrMap) Len() int {
	return len(m.ids)
}

// NewReversePtrMap returns an empty map from identifier to pointer.
func NewReversePtrMap() *ReversePtrMap {
	return &ReversePtrMap{
		ptrs: make(map[int]unsafe.Pointer),
	}
}

// ReversePtrMap resolves identifiers back to pointers while a value is decoded.
type ReversePtrMap struct {
	ptrs map[int]unsafe.Pointer
}

// Put stores ptr under key.
func (mp *ReversePtrMap) Put(key int, ptr unsafe.Pointer) {
	mp.ptrs[key] = ptr
}

// Get reports the pointer stored under key.
func (mp *ReversePtrMap) Get(key int) (ptr unsafe.Pointer, pst bool) {
	ptr, pst = mp.ptrs[key]
	return
}

// Len reports how many pointers have been stored.
func (mp *ReversePtrMap) Len() int {
	return len(mp.ptrs)
}
