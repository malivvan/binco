package binco

import (
	"strconv"
	"testing"
)

// TestWordLimits pins the wire width of int and uint: it is 64 bits on every
// host, so a payload written on amd64 reads back on a 32-bit build.
func TestWordLimits(t *testing.T) {
	if UintMaxVarintLen() != Uint64MaxVarintLen {
		t.Fatalf("UintMaxVarintLen() = %d, want %d (the width must not follow the host, which has %d-bit int)",
			UintMaxVarintLen(), Uint64MaxVarintLen, strconv.IntSize)
	}
	if UintMaxLastByte() != Uint64MaxLastByte {
		t.Fatalf("UintMaxLastByte() = %d, want %d", UintMaxLastByte(), Uint64MaxLastByte)
	}
	if IntWordBits != 64 {
		t.Fatalf("IntWordBits = %d, want 64", IntWordBits)
	}
	if IntWordBytes != 8 {
		t.Fatalf("IntWordBytes = %d, want 8", IntWordBytes)
	}
}
