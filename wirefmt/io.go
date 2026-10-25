package wirefmt

import (
	"bufio"
	"io"
)

// NewWriter turns any io.Writer into a Writer, wrapping it when it cannot write single bytes.
func NewWriter(w io.Writer) Writer {
	if bw, ok := w.(Writer); ok {
		return bw
	}
	return &writerAdapter{w: w}
}

// NewReader turns any io.Reader into a Reader, wrapping it when it cannot read single bytes.
func NewReader(r io.Reader) Reader {
	if br, ok := r.(Reader); ok {
		return br
	}
	return &readerAdapter{r: r}
}

// writerAdapter adds the byte and string writes that an io.Writer lacks.
type writerAdapter struct {
	w io.Writer
}

func (w *writerAdapter) WriteByte(c byte) error {
	_, err := w.w.Write([]byte{c})
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
	r io.Reader
}

func (r *readerAdapter) ReadByte() (b byte, err error) {
	var bs [1]byte
	if _, err = io.ReadFull(r.r, bs[:]); err != nil {
		return
	}
	return bs[0], nil
}

func (r *readerAdapter) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

// Encode writes v to w, buffering the writes when w cannot write single bytes.
func Encode[T any](s Serializer[T], w io.Writer, v T) (n int, err error) {
	var bw Writer

	var buf *bufio.Writer

	if sw, ok := w.(Writer); ok {
		bw = sw
	} else {
		buf = bufio.NewWriter(w)
		bw = buf
	}
	n, err = s.Marshal(v, bw)
	if buf != nil {
		if ferr := buf.Flush(); err == nil {
			err = ferr
		}
	}
	return
}

// Decode reads a value from r, buffering the reads when r cannot read single bytes.
func Decode[T any](s Serializer[T], r io.Reader) (v T, n int, err error) {
	var br Reader

	var buf *bufio.Reader

	if sr, ok := r.(Reader); ok {
		br = sr
	} else {
		buf = bufio.NewReader(r)
		br = buf
	}
	return s.Unmarshal(br)
}
