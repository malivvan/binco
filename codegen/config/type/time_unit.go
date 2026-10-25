package tpopts

import "fmt"

// The units a timestamp can be stored in.
const (
	TimeUnitUndefined TimeUnit = iota

	TimeUnitSec

	TimeUnitMilli

	TimeUnitMicro

	TimeUnitNano

	TimeUnitSecUTC

	TimeUnitMilliUTC

	TimeUnitMicroUTC

	TimeUnitNanoUTC
)

// TimeUnit selects the unit and the UTC handling of a timestamp.
type TimeUnit int

// Serializer names the time codec of the unit.
func (u TimeUnit) Serializer() string {
	switch u {
	case TimeUnitSec:
		return "TimeUnix"
	case TimeUnitMilli:
		return "TimeUnixMilli"
	case TimeUnitMicro:
		return "TimeUnixMicro"
	case TimeUnitNano:
		return "TimeUnixNano"
	case TimeUnitSecUTC:
		return "TimeUnixUTC"
	case TimeUnitMilliUTC:
		return "TimeUnixMilliUTC"
	case TimeUnitMicroUTC:
		return "TimeUnixMicroUTC"
	case TimeUnitNanoUTC:
		return "TimeUnixNanoUTC"
	default:
		panic(fmt.Sprintf("unexpected TimeUnit %v", u))
	}
}
