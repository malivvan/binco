package zerocpy

import (
	"unsafe"

	"github.com/malivvan/binco"
)

// The package refuses to run on a big-endian host: the direct memory access below assumes little-endian byte order, and a wrong result there would be silent.
func init() {
	var probe uint16 = 1
	if *(*byte)(unsafe.Pointer(&probe)) != 1 {
		panic(binco.ErrUnsupportedArch)
	}
}
