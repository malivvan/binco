package zerocpy

import (
	"time"

	"github.com/malivvan/binco"
)

// TimeUnix serializes a time.Time as whole Unix seconds.
var TimeUnix = unixView{}

type unixView struct{}

func (s unixView) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.Unix(), bs)
}

func (s unixView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	sec, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.Unix(sec, 0)
	return
}

func (s unixView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s unixView) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixMilli serializes a time.Time as Unix milliseconds.
var TimeUnixMilli = milliView{}

type milliView struct{}

func (s milliView) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.UnixMilli(), bs)
}

func (s milliView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	milli, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.UnixMilli(milli)
	return
}

func (s milliView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s milliView) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixMicro serializes a time.Time as Unix microseconds.
var TimeUnixMicro = microView{}

type microView struct{}

func (s microView) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.UnixMicro(), bs)
}

func (s microView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	micro, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.UnixMicro(micro)
	return
}

func (s microView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s microView) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixNano serializes a time.Time as Unix nanoseconds.
var TimeUnixNano = nanoView{}

type nanoView struct{}

func (s nanoView) Marshal(v time.Time, bs []byte) (n int) {
	return Int64.Marshal(v.UnixNano(), bs)
}

func (s nanoView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	nano, n, err := Int64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = time.Unix(0, nano)
	return
}

func (s nanoView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s nanoView) Skip(bs []byte) (n int, err error) {
	return Int64.Skip(bs)
}

// TimeUnixUTC is TimeUnix, but the decoded time is always in UTC.
var TimeUnixUTC = unixUTCView{}

type unixUTCView struct {
	unixView
}

func (s unixUTCView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.unixView.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}

// TimeUnixMilliUTC is TimeUnixMilli, but the decoded time is always in UTC.
var TimeUnixMilliUTC = milliUTCView{}

type milliUTCView struct {
	milliView
}

func (s milliUTCView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.milliView.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}

// TimeUnixMicroUTC is TimeUnixMicro, but the decoded time is always in UTC.
var TimeUnixMicroUTC = microUTCView{}

type microUTCView struct {
	microView
}

func (s microUTCView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.microView.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}

// TimeUnixNanoUTC is TimeUnixNano, but the decoded time is always in UTC.
var TimeUnixNanoUTC = nanoUTCView{}

type nanoUTCView struct {
	nanoView
}

func (s nanoUTCView) Unmarshal(bs []byte) (v time.Time, n int, err error) {
	v, n, err = s.nanoView.Unmarshal(bs)
	if err == nil {
		v = v.UTC()
	}
	return
}
