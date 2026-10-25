package binco

import (
	"errors"
	"math"
	"strconv"
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

// TestWordNarrowing checks the host-width conversion used by every int and uint
// codec. The width is an argument, so the 32-bit rule is testable on a 64-bit
// host; the rows that need a 64-bit host only run there.
func TestWordNarrowing(t *testing.T) {
	checkInt := func(t *testing.T, raw int64, width int, want int64, wantErr error) {
		t.Helper()
		got, err := WordInt(raw, width)
		if !errors.Is(err, wantErr) {
			t.Fatalf("WordInt(%d, %d) error = %v, want %v", raw, width, err, wantErr)
		}
		if err == nil && int64(got) != want {
			t.Fatalf("WordInt = %d, want %d", got, want)
		}
	}
	checkUint := func(t *testing.T, raw uint64, width int, want uint64, wantErr error) {
		t.Helper()
		got, err := WordUint(raw, width)
		if !errors.Is(err, wantErr) {
			t.Fatalf("WordUint(%d, %d) error = %v, want %v", raw, width, err, wantErr)
		}
		if err == nil && uint64(got) != want {
			t.Fatalf("WordUint = %d, want %d", got, want)
		}
	}

	t.Run("int32", func(t *testing.T) {
		checkInt(t, math.MaxInt32, 32, math.MaxInt32, nil)
		checkInt(t, math.MinInt32, 32, math.MinInt32, nil)
		checkInt(t, math.MaxInt32+1, 32, 0, ErrOverflow)
		checkInt(t, math.MinInt32-1, 32, 0, ErrOverflow)
	})
	t.Run("uint32", func(t *testing.T) {
		checkUint(t, math.MaxUint32, 32, math.MaxUint32, nil)
		checkUint(t, math.MaxUint32+1, 32, 0, ErrOverflow)
	})
	if strconv.IntSize == 64 {
		t.Run("int64", func(t *testing.T) {
			checkInt(t, 1<<40, 64, 1<<40, nil)
			checkInt(t, math.MinInt64, 64, math.MinInt64, nil)
		})
		t.Run("uint64", func(t *testing.T) {
			checkUint(t, 1<<40, 64, 1<<40, nil)
			checkUint(t, math.MaxUint64, 64, math.MaxUint64, nil)
		})
	}
}
