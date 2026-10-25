package le_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strconv"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/le"
	"github.com/malivvan/binco/internal/lew"
)

// TestBufferAndStreamAgree checks the two implementations against each other:
// the bytes internal/le writes must be the bytes internal/lew reads, and the
// other way around.
func TestBufferAndStreamAgree(t *testing.T) {
	t.Run("64", func(t *testing.T) {
		var buf bytes.Buffer
		if _, err := lew.Write64(&buf, int64(-1234567890)); err != nil {
			t.Fatalf("Write64: %v", err)
		}
		written := make([]byte, 8)
		le.Put64(written, int64(-1234567890))
		if !bytes.Equal(buf.Bytes(), written) {
			t.Fatalf("stream %v, buffer %v", buf.Bytes(), written)
		}
		fromBuffer, n, err := lew.Read64[int64](bytes.NewReader(written))
		if err != nil || n != 8 || fromBuffer != -1234567890 {
			t.Fatalf("Read64 = (%d, %d, %v)", fromBuffer, n, err)
		}
		fromStream, n, err := le.Get64[int64](buf.Bytes())
		if err != nil || n != 8 || fromStream != -1234567890 {
			t.Fatalf("Get64 = (%d, %d, %v)", fromStream, n, err)
		}
	})

	t.Run("32 unaligned", func(t *testing.T) {
		bs := make([]byte, 5)
		le.Put32(bs[1:], uint32(0x01020304))
		if got, _, err := le.Get32[uint32](bs[1:]); err != nil || got != 0x01020304 {
			t.Fatalf("Get32 at an odd offset = (%#x, %v)", got, err)
		}
		if got := bs[1:]; !bytes.Equal(got, []byte{4, 3, 2, 1}) {
			t.Fatalf("Put32 = %v", got)
		}
	})

	t.Run("16", func(t *testing.T) {
		var buf bytes.Buffer
		if _, err := lew.Write16(&buf, uint16(0x0102)); err != nil {
			t.Fatalf("Write16: %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{2, 1}) {
			t.Fatalf("Write16 = %v", buf.Bytes())
		}
		if got, _, err := le.Get16[uint16](buf.Bytes()); err != nil || got != 0x0102 {
			t.Fatalf("Get16 = (%#x, %v)", got, err)
		}
	})

	t.Run("8", func(t *testing.T) {
		if got := le.Put8(make([]byte, 1), int8(-1)); got != 1 {
			t.Fatalf("Put8 size = %d", got)
		}
		if got, _, err := le.Get8[uint8]([]byte{0xff}); err != nil || got != 0xff {
			t.Fatalf("Get8 = (%#x, %v)", got, err)
		}
	})
}

// TestWordWidthIs64 checks that int and uint use the same 64-bit width on every
// host, and that a value the host cannot hold is ErrOverflow.
func TestWordWidthIs64(t *testing.T) {
	bs := make([]byte, 8)
	if n := le.PutWord(bs, 300); n != 8 {
		t.Fatalf("PutWord wrote %d bytes, want 8", n)
	}
	if got, n, err := le.GetWord(bs); err != nil || n != 8 || got != 300 {
		t.Fatalf("GetWord = (%d, %d, %v)", got, n, err)
	}

	var buf bytes.Buffer
	if _, err := lew.WriteWord(&buf, -1); err != nil {
		t.Fatalf("WriteWord: %v", err)
	}
	if got, _, err := lew.ReadWord(&buf); err != nil || got != -1 {
		t.Fatalf("ReadWord = (%d, %v)", got, err)
	}

	// The wire value 2^63 does not fit a 32-bit int. On a 64-bit host it is
	// MinInt64 and decodes without error, so the overflow check only has a case
	// to report there.
	big := make([]byte, 8)
	le.Put64(big, uint64(1)<<63)
	if strconv.IntSize == 64 {
		if got, _, err := le.GetWord(big); err != nil || int64(got) != math.MinInt64 {
			t.Fatalf("GetWord of 2^63 on a 64-bit host = (%d, %v), want MinInt64", got, err)
		}
		return
	}
	if _, _, err := le.GetWord(big); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("GetWord of 2^63 returned %v, want ErrOverflow", err)
	}
	if _, err := le.SkipWord(big); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("SkipWord of 2^63 returned %v, want ErrOverflow", err)
	}
	if _, err := lew.SkipWord(bytes.NewReader(big)); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("stream SkipWord of 2^63 returned %v, want ErrOverflow", err)
	}
}

// TestUnsignedWord checks the unsigned helpers on both sides.
func TestUnsignedWord(t *testing.T) {
	const value = uint(1) << 20
	bs := make([]byte, 8)
	le.PutWordUnsigned(bs, value)
	if got, _, err := le.GetWordUnsigned(bs); err != nil || got != value {
		t.Fatalf("GetWordUnsigned = (%d, %v)", got, err)
	}

	var buf bytes.Buffer
	if _, err := lew.WriteWordUnsigned(&buf, value); err != nil {
		t.Fatalf("WriteWordUnsigned: %v", err)
	}
	if got, _, err := lew.ReadWordUnsigned(&buf); err != nil || got != value {
		t.Fatalf("ReadWordUnsigned = (%d, %v)", got, err)
	}
}

// TestShortInput checks that every reader reports the missing bytes.
func TestShortInput(t *testing.T) {
	if _, _, err := le.Get64[uint64](make([]byte, 7)); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Get64 = %v", err)
	}
	if _, err := le.Skip32(make([]byte, 3)); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Skip32 = %v", err)
	}
	// A stream distinguishes a clean end (io.EOF, nothing read) from a value
	// that started and was cut short (io.ErrUnexpectedEOF).
	if _, _, err := lew.Read16[uint16](bytes.NewReader(make([]byte, 1))); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("stream Read16 = %v, want io.ErrUnexpectedEOF", err)
	}
	if _, err := lew.Skip64(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
		t.Fatalf("stream Skip64 = %v, want io.EOF", err)
	}
}
