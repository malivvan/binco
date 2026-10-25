package wirefmt_test

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
)

// plainReader hides the ReadByte method of the reader it wraps, which is what a
// net.Conn, an os.File or an io.Pipe look like to this package.
type plainReader struct{ r io.Reader }

func (p plainReader) Read(b []byte) (int, error) { return p.r.Read(b) }

// TestDecodeDoesNotSwallowTheNextMessage covers S7: Decode used to wrap a plain
// reader in a bufio.Reader per call, read ahead, and drop the next message.
func TestDecodeDoesNotSwallowTheNextMessage(t *testing.T) {
	var first, second bytes.Buffer
	if _, err := wirefmt.Encode(wcomposite.String, &first, "hello"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := wirefmt.Encode(wcomposite.String, &second, "world"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	stream := append(append([]byte{}, first.Bytes()...), second.Bytes()...)

	r := plainReader{bytes.NewReader(stream)}
	got, n, err := wirefmt.Decode(wcomposite.String, r)
	if err != nil || got != "hello" || n != first.Len() {
		t.Fatalf("first Decode = (%q, %d, %v)", got, n, err)
	}
	got, n, err = wirefmt.Decode(wcomposite.String, r)
	if err != nil || got != "world" || n != second.Len() {
		t.Fatalf("second Decode = (%q, %d, %v)", got, n, err)
	}
	if _, _, err = wirefmt.Decode(wcomposite.String, r); !errors.Is(err, io.EOF) {
		t.Fatalf("third Decode = %v, want io.EOF", err)
	}
}

// TestDecoderReusesOneBuffer covers the stateful replacement for the old
// per-call bufio: a Decoder can read every message of a stream.
func TestDecoderReusesOneBuffer(t *testing.T) {
	var buf bytes.Buffer
	for _, v := range []string{"a", "bb", "ccc"} {
		if _, err := wirefmt.Encode(wcomposite.String, &buf, v); err != nil {
			t.Fatalf("Encode: %v", err)
		}
	}
	d := wirefmt.NewDecoder(&buf)
	for _, want := range []string{"a", "bb", "ccc"} {
		got, _, err := wirefmt.DecodeFrom(d, wcomposite.String)
		if err != nil || got != want {
			t.Fatalf("DecodeFrom = (%q, %v), want %q", got, err, want)
		}
	}
}

// TestTruncatedStreamIsUnexpectedEOF covers the EOF mapping of S7: only the
// first byte of a value decides between a clean end and a truncated message.
func TestTruncatedStreamIsUnexpectedEOF(t *testing.T) {
	r := func(bs ...byte) wirefmt.Reader { return wirefmt.NewReader(bytes.NewReader(bs)) }

	if _, _, err := wcomposite.String.Unmarshal(r()); !errors.Is(err, io.EOF) {
		t.Fatalf("empty stream returned %v, want io.EOF", err)
	}
	if _, _, err := wcomposite.String.Unmarshal(r(5, 'a')); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated string returned %v, want io.ErrUnexpectedEOF", err)
	}
	if _, _, err := wvarnum.Uint64.Unmarshal(r(0x80)); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated varint returned %v, want io.ErrUnexpectedEOF", err)
	}
}

// TestStreamingHugeLengthIsBounded covers S1 and S2 for the stream codecs: a
// huge declared length must fail without allocating it in advance.
func TestStreamingHugeLengthIsBounded(t *testing.T) {
	maxInt64 := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}

	before := allocBytes()
	if _, _, err := wcomposite.String.Unmarshal(wirefmt.NewReader(bytes.NewReader(maxInt64))); err == nil {
		t.Fatal("truncated string returned no error")
	}
	if _, _, err := wcomposite.ByteSlice.Unmarshal(wirefmt.NewReader(bytes.NewReader(maxInt64))); err == nil {
		t.Fatal("truncated blob returned no error")
	}
	if _, _, err := wcomposite.NewSliceSerializer(wvarnum.Int).Unmarshal(
		wirefmt.NewReader(bytes.NewReader([]byte{0x80, 0x80, 0x80, 0x80, 0x01}))); err == nil {
		t.Fatal("truncated slice returned no error")
	}
	if grew := allocBytes() - before; grew > 8<<20 {
		t.Fatalf("decoding three truncated streams allocated %d bytes", grew)
	}
}

func allocBytes() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.TotalAlloc
}

// TestStreamingZeroSizeElementsAreBounded covers S10 for the stream codecs.
func TestStreamingZeroSizeElementsAreBounded(t *testing.T) {
	var buf bytes.Buffer
	if _, err := wvarnum.PositiveInt.Marshal(binco.MaxZeroSizeElements+1, &buf); err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if _, _, err := wcomposite.NewSliceSerializer(zeroSerializer{}).Unmarshal(
		wirefmt.NewReader(&buf)); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("Unmarshal returned %v, want ErrTooLargeLength", err)
	}
}

// zeroSerializer encodes a value that occupies no bytes at all.
type zeroSerializer struct{}

func (zeroSerializer) Marshal(v struct{}, w wirefmt.Writer) (n int, err error) {
	return 0, nil
}
func (zeroSerializer) Unmarshal(r wirefmt.Reader) (struct{}, int, error) { return struct{}{}, 0, nil }
func (zeroSerializer) Size(v struct{}) (size int)                        { return 0 }
func (zeroSerializer) Skip(r wirefmt.Reader) (n int, err error)          { return 0, nil }
