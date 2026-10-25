package scalar

import (
	"time"

	"github.com/malivvan/binco"
)

// TimeUnix serializes a time.Time as whole Unix seconds.
var TimeUnix = unixCodec{}

type unixCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s unixCodec) FixedWidth() int { return binco.Num64RawSize }

func (s unixCodec) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.Unix(), bs)
}

func (s unixCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	sec, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.Unix(sec, 0).UTC()
	return
}

func (s unixCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s unixCodec) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixMilli serializes a time.Time as Unix milliseconds.
var TimeUnixMilli = milliCodec{}

type milliCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s milliCodec) FixedWidth() int { return binco.Num64RawSize }

func (s milliCodec) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.UnixMilli(), bs)
}

func (s milliCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	milli, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.UnixMilli(milli).UTC()
	return
}

func (s milliCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s milliCodec) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixMicro serializes a time.Time as Unix microseconds.
var TimeUnixMicro = microCodec{}

type microCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s microCodec) FixedWidth() int { return binco.Num64RawSize }

func (s microCodec) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.UnixMicro(), bs)
}

func (s microCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	micro, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.UnixMicro(micro).UTC()
	return
}

func (s microCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s microCodec) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixNano serializes a time.Time as Unix nanoseconds.
var TimeUnixNano = nanoCodec{}

type nanoCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s nanoCodec) FixedWidth() int { return binco.Num64RawSize }

func (s nanoCodec) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.UnixNano(), bs)
}

func (s nanoCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	nano, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.Unix(0, nano).UTC()
	return
}

func (s nanoCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s nanoCodec) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixUTC is TimeUnix, but the decoded time is always in UTC.
var TimeUnixUTC = unixUTCCodec{}

type unixUTCCodec struct {
	unixCodec
}

func (s unixUTCCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.unixCodec.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}

// TimeUnixMilliUTC is TimeUnixMilli, but the decoded time is always in UTC.
var TimeUnixMilliUTC = milliUTCCodec{}

type milliUTCCodec struct {
	milliCodec
}

func (s milliUTCCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.milliCodec.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}

// TimeUnixMicroUTC is TimeUnixMicro, but the decoded time is always in UTC.
var TimeUnixMicroUTC = microUTCCodec{}

type microUTCCodec struct {
	microCodec
}

func (s microUTCCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.microCodec.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}

// TimeUnixNanoUTC is TimeUnixNano, but the decoded time is always in UTC.
var TimeUnixNanoUTC = nanoUTCCodec{}

type nanoUTCCodec struct {
	nanoCodec
}

func (s nanoUTCCodec) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.nanoCodec.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}
