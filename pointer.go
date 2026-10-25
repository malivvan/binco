package binco

import (
	"reflect"
	"unsafe"
)

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

// PtrKey identifies a pointer together with the type of the value it points to.
//
// The type is part of the key on purpose. Two pointers to different types can
// share an address (a struct and its first field, or any two zero-size values),
// and a pointer table keyed by the address alone would then hand the second
// value the identifier of the first one and decode it as the wrong type.
type PtrKey struct {
	Ptr  unsafe.Pointer
	Type reflect.Type
}

// NewPtrMap returns an empty map from pointer to identifier.
func NewPtrMap() *PtrMap {
	return &PtrMap{
		ids: make(map[PtrKey]int),
	}
}

// PtrMap hands out increasing identifiers for pointers while a value is encoded.
//
// A PtrMap is stateful: it is reset by PtrMap.Clear (or by ptrmap.Wrap) between
// values and is not safe for concurrent use.
type PtrMap struct {
	next int
	ids  map[PtrKey]int
}

// Put stores the pointer and its type and returns the identifier assigned to
// them.
func (m *PtrMap) Put(ptr unsafe.Pointer, typ reflect.Type) (key int) {
	key = m.next
	m.ids[PtrKey{Ptr: ptr, Type: typ}] = key
	m.next++
	return
}

// Get reports the identifier of the pointer, if this exact pointer and type
// have been stored.
func (m *PtrMap) Get(ptr unsafe.Pointer, typ reflect.Type) (key int, pst bool) {
	key, pst = m.ids[PtrKey{Ptr: ptr, Type: typ}]
	return
}

// Len reports how many pointers have been stored.
func (m *PtrMap) Len() int {
	return len(m.ids)
}

// Next reports the identifier the next Put will assign.
func (m *PtrMap) Next() int {
	return m.next
}

// Clear forgets every pointer and restarts the identifiers at zero.
func (m *PtrMap) Clear() {
	clear(m.ids)
	m.next = 0
}

// PtrEntry is one entry of a ReversePtrMap: the pointer that was decoded and
// the type it was decoded as.
type PtrEntry struct {
	Ptr  unsafe.Pointer
	Type reflect.Type
}

// NewReversePtrMap returns an empty map from identifier to pointer.
func NewReversePtrMap() *ReversePtrMap {
	return &ReversePtrMap{
		ptrs: make(map[int]PtrEntry),
	}
}

// ReversePtrMap resolves identifiers back to pointers while a value is decoded.
//
// Like PtrMap it is stateful and not safe for concurrent use.
type ReversePtrMap struct {
	ptrs  map[int]PtrEntry
	depth Depth
}

// Put stores ptr under key together with the type it was decoded as.
func (mp *ReversePtrMap) Put(key int, ptr unsafe.Pointer, typ reflect.Type) {
	mp.ptrs[key] = PtrEntry{Ptr: ptr, Type: typ}
}

// Get reports the pointer stored under key, the type it was decoded as, and
// whether the key is known.
func (mp *ReversePtrMap) Get(key int) (ptr unsafe.Pointer, typ reflect.Type, pst bool) {
	entry, pst := mp.ptrs[key]
	return entry.Ptr, entry.Type, pst
}

// Len reports how many pointers have been stored.
func (mp *ReversePtrMap) Len() int {
	return len(mp.ptrs)
}

// Enter records one level of reference nesting.
//
// The counter lives on the map because every codec in one decode shares it, so
// the limit applies to the whole value rather than to each codec separately.
func (mp *ReversePtrMap) Enter() error {
	return mp.depth.Enter()
}

// Leave ends one level of reference nesting.
func (mp *ReversePtrMap) Leave() {
	mp.depth.Leave()
}

// Clear forgets every pointer and resets the nesting counter.
func (mp *ReversePtrMap) Clear() {
	clear(mp.ptrs)
	mp.depth.Reset()
}
