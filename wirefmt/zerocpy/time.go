package zerocpy

import (
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// TimeUnix serializes a time.Time as whole Unix seconds.
var TimeUnix = unixView{}

type unixView struct{}

func (s unixView) Marshal(v time.Time, w wirefmt.Writer) (n int, err error) {
	return Int64.Marshal(v.Unix(), w)
}

func (s unixView) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	sec, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.Unix(sec, 0)
	return
}

func (s unixView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s unixView) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixMilli serializes a time.Time as Unix milliseconds.
var TimeUnixMilli = milliView{}

type milliView struct{}

func (s milliView) Marshal(v time.Time, w wirefmt.Writer) (n int,
	err error,
) {
	return Int64.Marshal(v.UnixMilli(), w)
}

func (s milliView) Unmarshal(r wirefmt.Reader) (v time.Time, n int,
	err error,
) {
	milli, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.UnixMilli(milli)
	return
}

func (s milliView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s milliView) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixMicro serializes a time.Time as Unix microseconds.
var TimeUnixMicro = microView{}

type microView struct{}

func (s microView) Marshal(v time.Time, w wirefmt.Writer) (n int,
	err error,
) {
	return Int64.Marshal(v.UnixMicro(), w)
}

func (s microView) Unmarshal(r wirefmt.Reader) (v time.Time, n int,
	err error,
) {
	micro, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.UnixMicro(micro)
	return
}

func (s microView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s microView) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixNano serializes a time.Time as Unix nanoseconds.
var TimeUnixNano = nanoView{}

type nanoView struct{}

func (s nanoView) Marshal(v time.Time, w wirefmt.Writer) (n int, err error) {
	return Int64.Marshal(v.UnixNano(), w)
}

func (s nanoView) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	nano, n, err := Int64.Unmarshal(r)
	if err != nil {
		return
	}
	v = time.Unix(0, nano)
	return
}

func (s nanoView) Size(v time.Time) (size int) {
	return binco.Num64RawSize
}

func (s nanoView) Skip(r wirefmt.Reader) (n int, err error) {
	return Int64.Skip(r)
}

// TimeUnixUTC is TimeUnix, but the decoded time is always in UTC.
var TimeUnixUTC = unixUTCView{}

type unixUTCView struct {
	unixView
}

func (s unixUTCView) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.unixView.Unmarshal(r)
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

func (s milliUTCView) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.milliView.Unmarshal(r)
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

func (s microUTCView) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.microView.Unmarshal(r)
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

func (s nanoUTCView) Unmarshal(r wirefmt.Reader) (v time.Time, n int, err error) {
	v, n, err = s.nanoView.Unmarshal(r)
	if err == nil {
		v = v.UTC()
	}
	return
}
