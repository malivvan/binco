package testdata

// SameName exists twice in this module, here and in testdata/other, so that the
// generator's name-collision check has something to reject.
type SameName struct {
	Value int
}
