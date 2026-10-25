package binco

import (
	"testing"
)

// TestWordLimits pins the wire width of int and uint: it is 64 bits on every
// host, so a payload written on amd64 reads back on a 32-bit build.
func TestWordLimits(t *testing.T) {
	if IntWordBits != 64 {
		t.Fatalf("IntWordBits = %d, want 64", IntWordBits)
	}
	if IntWordBytes != 8 {
		t.Fatalf("IntWordBytes = %d, want 8", IntWordBytes)
	}
}
