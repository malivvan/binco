package binco_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/malivvan/binco"
)

// TestReversePtrMapScope checks the decode-side map: lookups report the type an
// id was defined with, Leave ends a level, and Clear forgets both the pointers
// and the nesting budget.
func TestReversePtrMapScope(t *testing.T) {
	rev := binco.NewReversePtrMap()
	v := 5
	ptr := unsafe.Pointer(&v)
	intPtr := reflect.TypeFor[*int]()
	rev.Put(1, ptr, intPtr)

	got, typ, ok := rev.Get(1)
	if !ok || got != ptr || typ != intPtr {
		t.Fatalf("Get(1) = (%v, %v, %v), want (%v, %v, true)", got, typ, ok, ptr, intPtr)
	}
	if rev.Len() != 1 {
		t.Fatalf("Len = %d, want 1", rev.Len())
	}

	if err := rev.Enter(); err != nil {
		t.Fatalf("Enter: %v", err)
	}
	rev.Leave()

	rev.Clear()
	if rev.Len() != 0 {
		t.Fatalf("Clear left %d entries", rev.Len())
	}
	if _, _, ok := rev.Get(1); ok {
		t.Fatal("Get after Clear reported the entry")
	}

	// Clear resets the depth counter, so a full budget is available again.
	for range binco.MaxDepth {
		if err := rev.Enter(); err != nil {
			t.Fatalf("Enter after Clear: %v", err)
		}
	}
	rev.Clear()
	if err := rev.Enter(); err != nil {
		t.Fatalf("Enter after the second Clear: %v", err)
	}
}
