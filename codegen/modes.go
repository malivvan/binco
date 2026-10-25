package codegen

import (
	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

// The package names the generated code refers to.
const (
	VarnumPkg          = "varnum"
	ScalarPkg          = "scalar"
	CompositePkg       = "composite"
	ZerocopyPkg        = "zerocopy"
	DefTimeSerializer  = "TimeUnixUTC"
	TimeSerializerRule = "time_serializer"
)

// portableRules maps a Go type to the package that encodes it in safe mode.
func portableRules(tops tpopts.Options) map[string]string {
	var (
		m              = make(map[string]string)
		numPkg         = VarnumPkg
		timeSerializer = DefTimeSerializer
	)
	if tops.NumEncoding != tpopts.NumEncodingUndefined {
		numPkg = tops.NumEncoding.Package()
	}
	if tops.TimeUnit != tpopts.TimeUnitUndefined {
		timeSerializer = tops.TimeUnit.Serializer()
	}
	m["int"] = numPkg
	m["int64"] = numPkg
	m["int32"] = numPkg
	m["int16"] = numPkg
	m["uint"] = numPkg
	m["uint64"] = numPkg
	m["uint32"] = numPkg
	m["uint16"] = numPkg

	m["int8"] = ScalarPkg
	m["uint8"] = ScalarPkg
	m["float64"] = ScalarPkg
	m["float32"] = ScalarPkg
	m["byte"] = ScalarPkg
	m["time.Time"] = ScalarPkg

	m["bool"] = CompositePkg
	m["string"] = CompositePkg
	m["[]byte"] = CompositePkg
	m["[]uint8"] = CompositePkg

	m[TimeSerializerRule] = timeSerializer
	return m
}

// nativeRules maps a Go type to the package that encodes it in unsafe mode.
func nativeRules(tops tpopts.Options, opts genopts.Options) map[string]string {
	var (
		m              = make(map[string]string)
		numPkg         = ZerocopyPkg
		byteSlicePkg   = ZerocopyPkg
		timeSerializer = DefTimeSerializer
	)
	if tops.NumEncoding != tpopts.NumEncodingUndefined {
		numPkg = tops.NumEncoding.Package()
	}
	if tops.TimeUnit != tpopts.TimeUnitUndefined {
		timeSerializer = tops.TimeUnit.Serializer()
	}
	if opts.Stream {
		byteSlicePkg = CompositePkg
	}
	m["int"] = numPkg
	m["int64"] = numPkg
	m["int32"] = numPkg
	m["int16"] = numPkg
	m["uint"] = numPkg
	m["uint64"] = numPkg
	m["uint32"] = numPkg
	m["uint16"] = numPkg

	m["int8"] = ZerocopyPkg
	m["uint8"] = ZerocopyPkg
	m["float64"] = ZerocopyPkg
	m["float32"] = ZerocopyPkg
	m["time.Time"] = ZerocopyPkg
	m["byte"] = ZerocopyPkg
	m["bool"] = ZerocopyPkg
	m["string"] = ZerocopyPkg

	m["[]byte"] = byteSlicePkg
	m["[]uint8"] = byteSlicePkg

	m[TimeSerializerRule] = timeSerializer
	return m
}

// hybridRules is nativeRules, except that strings keep the portable codec.
func hybridRules(tops tpopts.Options, opts genopts.Options) map[string]string {
	m := nativeRules(tops, opts)
	m["string"] = CompositePkg
	return m
}
