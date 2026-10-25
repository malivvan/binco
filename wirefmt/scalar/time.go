package scalar

import (
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// TimeUnix serializes a time.Time as whole Unix seconds.
var TimeUnix = unixCodec{}

type unixCodec struct{}

func (s unixCodec) Marshal(v time.Time, w wirefmt.Writer) (n int, err error) {
	return Int64.Marshal(v.Unix(), w)
}

func (s unixCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	sec, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.Unix(sec, 0)
	return
}

func (s unixCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s unixCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixMilli serializes a time.Time as Unix milliseconds.
var TimeUnixMilli = milliCodec{}

type milliCodec struct{}

func (s milliCodec) Marshal(v time.Time, w wirefmt.Writer) (n int,
	err error,
) {
	return Int64.Marshal(v.UnixMilli(), w)
}

func (s milliCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int,
	err error,
) {
	milli, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.UnixMilli(milli)
	return
}

func (s milliCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s milliCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixMicro serializes a time.Time as Unix microseconds.
var TimeUnixMicro = microCodec{}

type microCodec struct{}

func (s microCodec) Marshal(v time.Time, w wirefmt.Writer) (n int,
	err error,
) {
	return Int64.Marshal(v.UnixMicro(), w)
}

func (s microCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int,
	err error,
) {
	micro, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.UnixMicro(micro)
	return
}

func (s microCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s microCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixNano serializes a time.Time as Unix nanoseconds.
var TimeUnixNano = nanoCodec{}

type nanoCodec struct{}

func (s nanoCodec) Marshal(v time.Time, w wirefmt.Writer) (n int, err error) {
	return Int64.Marshal(v.UnixNano(), w)
}

func (s nanoCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	nano, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.Unix(0, nano)
	return
}

func (s nanoCodec) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s nanoCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixUTC is TimeUnix, but the decoded time is always in UTC.
var TimeUnixUTC = unixUTCCodec{}

type unixUTCCodec struct {
	unixCodec
}

func (s unixUTCCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.unixCodec.Unmarshal(r)
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

func (s milliUTCCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.milliCodec.Unmarshal(r)
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

func (s microUTCCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.microCodec.Unmarshal(r)
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

func (s nanoUTCCodec) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.nanoCodec.Unmarshal(r)
	if err == nil {
		v = v.UTC()
	}
	return
}
