package wirefmt

import (
	"bufio"
	"io"
	"sync"

	"github.com/malivvan/binco/internal/wireio"
)

// writerPool recycles the bufio.Writer that Encode needs when the destination
// cannot write single bytes. Allocating one per call cost 4 KiB per message.
var writerPool = sync.Pool{
	New: func() any { return bufio.NewWriterSize(io.Discard, 4096) },
}

// NewWriter turns any io.Writer into a Writer, wrapping it when it cannot write single bytes.
//
// The wrapper writes one byte per call and keeps its scratch byte in the
// adapter, so it does not allocate per byte. It never buffers ahead: bytes it
// does not write are still in the underlying writer, which matters when several
// values share one stream.
func NewWriter(w io.Writer) Writer {
	if bw, ok := w.(Writer); ok {
		return bw
	}
	return &writerAdapter{w: w}
}

// NewReader turns any io.Reader into a Reader, wrapping it when it cannot read single bytes.
//
// The wrapper reads exactly one byte per call and never reads ahead, so a
// reader that is shared between messages does not lose the bytes of the next
// one. Use NewDecoder when the codec is the only consumer of the stream: it
// buffers and is considerably faster.
func NewReader(r io.Reader) Reader {
	if br, ok := r.(Reader); ok {
		return br
	}
	return &readerAdapter{r: r}
}

// writerAdapter adds the byte and string writes that an io.Writer lacks.
type writerAdapter struct {
	w   io.Writer
	buf [1]byte
}

func (w *writerAdapter) WriteByte(c byte) error {
	w.buf[0] = c
	_, err := w.w.Write(w.buf[:])
	return err
}

func (w *writerAdapter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writerAdapter) WriteString(s string) (n int, err error) {
	return io.WriteString(w.w, s)
}

// readerAdapter adds the single-byte read that an io.Reader lacks.
type readerAdapter struct {
	r   io.Reader
	buf [1]byte
}

func (r *readerAdapter) ReadByte() (b byte, err error) {
	// wireio.ReadFull, not io.ReadFull: a reader that keeps returning (0, nil)
	// must fail with io.ErrNoProgress instead of spinning.
	if _, err = wireio.ReadFull(r.r, r.buf[:]); err != nil {
		return
	}
	return r.buf[0], nil
}

func (r *readerAdapter) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

// Encoder owns one buffered writer for the lifetime of a stream.
//
// Encode allocates a bufio.Writer per call; an Encoder keeps one, so a loop that
// writes many messages allocates nothing after the first. An Encoder is not safe
// for concurrent use.
type Encoder struct {
	bw *bufio.Writer
	w  Writer
}

// NewEncoder returns an Encoder that writes to w.
//
// A w that already implements Writer (a *bufio.Writer, for example) is used
// directly; anything else is wrapped in one.
func NewEncoder(w io.Writer) *Encoder {
	if bw, ok := w.(*bufio.Writer); ok {
		return &Encoder{bw: bw, w: bw}
	}
	if sw, ok := w.(Writer); ok {
		return &Encoder{w: sw}
	}
	bw := bufio.NewWriter(w)
	return &Encoder{bw: bw, w: bw}
}

// Writer returns the buffered writer the codecs write to.
func (e *Encoder) Writer() Writer {
	return e.w
}

// Flush writes everything that is still buffered.
func (e *Encoder) Flush() error {
	if e.bw == nil {
		return nil
	}
	return e.bw.Flush()
}

// Reset points the Encoder at w. Bytes that are still buffered are dropped
// without being written, so flush first when they matter.
func (e *Encoder) Reset(w io.Writer) {
	if e.bw != nil {
		e.bw.Reset(w)
		return
	}
	*e = *NewEncoder(w)
}

// Decoder owns one buffered reader for the lifetime of a stream.
//
// Decode does not buffer when the reader cannot read single bytes, because it
// cannot know whether the caller still needs the bytes it would read ahead.
// A Decoder is the right choice when the codec reads the stream to its end; it
// buffers once and reuses the buffer for every message.
//
// A Decoder is not safe for concurrent use.
type Decoder struct {
	br *bufio.Reader
	r  Reader
}

// NewDecoder returns a Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	if br, ok := r.(*bufio.Reader); ok {
		return &Decoder{br: br, r: br}
	}
	if sr, ok := r.(Reader); ok {
		return &Decoder{r: sr}
	}
	br := bufio.NewReader(r)
	return &Decoder{br: br, r: br}
}

// Reader returns the buffered reader the codecs read from.
func (d *Decoder) Reader() Reader {
	return d.r
}

// Reset points the Decoder at r. Bytes that the previous reader had buffered
// but the codec did not consume are dropped, so reset only at a message
// boundary.
func (d *Decoder) Reset(r io.Reader) {
	if d.br != nil {
		d.br.Reset(r)
		return
	}
	*d = *NewDecoder(r)
}

// Encode writes v to w.
//
// When w cannot write single bytes the writes are buffered for the duration of
// the call and flushed before Encode returns. The buffer is allocated per call;
// use an Encoder in a loop.
func Encode[T any](s Serializer[T], w io.Writer, v T) (n int, err error) {
	if sw, ok := w.(Writer); ok {
		return s.Marshal(v, sw)
	}
	bw := writerPool.Get().(*bufio.Writer)
	bw.Reset(w)
	n, err = s.Marshal(v, bw)
	if ferr := bw.Flush(); err == nil {
		err = ferr
	}
	bw.Reset(io.Discard)
	writerPool.Put(bw)
	return
}

// EncodeTo writes v through an Encoder that the caller reuses.
func EncodeTo[T any](e *Encoder, s Serializer[T], v T) (n int, err error) {
	return s.Marshal(v, e.Writer())
}

// Decode reads a value from r.
//
// A reader that cannot read single bytes is wrapped in an adapter that reads
// exactly one byte per call. That is slower than buffering, but it never
// consumes bytes that belong to the next value: buffering per call used to
// swallow the next message of a shared stream. Use a Decoder when the codec owns
// the stream.
func Decode[T any](s Serializer[T], r io.Reader) (v T, n int, err error) {
	return s.Unmarshal(NewReader(r))
}

// DecodeFrom reads one value through a Decoder that the caller reuses.
func DecodeFrom[T any](d *Decoder, s Serializer[T]) (v T, n int, err error) {
	return s.Unmarshal(d.Reader())
}
