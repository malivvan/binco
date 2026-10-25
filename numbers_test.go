package binco

import (
	"strconv"
	"testing"
)

func TestWordLimits(t *testing.T) {
	tests := []struct {
		name     string
		intSize  int
		wantLen  int
		wantLast byte
		wantOK   bool
	}{
		{"64-bit", 64, Uint64MaxVarintLen, Uint64MaxLastByte, true},
		{"32-bit", 32, Uint32MaxVarintLen, Uint32MaxLastByte, true},
		{"16-bit", 16, Uint16MaxVarintLen, Uint16MaxLastByte, true},
		{"8-bit", 8, Uint8MaxVarintLen, Uint8MaxLastByte, true},
		{"zero", 0, 0, 0, false},
		{"unsupported", 128, 0, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			maxLen, maxLast, ok := wordLimits(tc.intSize)
			if maxLen != tc.wantLen || maxLast != tc.wantLast || ok != tc.wantOK {
				t.Fatalf("wordLimits(%d) = (%d, %d, %v), want (%d, %d, %v)",
					tc.intSize, maxLen, maxLast, ok, tc.wantLen, tc.wantLast, tc.wantOK)
			}
		})
	}
}

func TestHostWordLimits(t *testing.T) {
	maxLen, maxLast, ok := wordLimits(strconv.IntSize)
	if !ok {
		t.Fatalf("wordLimits(%d) is not supported", strconv.IntSize)
	}
	if UintMaxVarintLen() != maxLen {
		t.Fatalf("UintMaxVarintLen() = %d, want %d", UintMaxVarintLen(), maxLen)
	}
	if UintMaxLastByte() != maxLast {
		t.Fatalf("UintMaxLastByte() = %d, want %d", UintMaxLastByte(), maxLast)
	}
}
