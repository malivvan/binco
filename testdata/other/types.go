// Package other is a second package with a type whose name also exists in
// testdata, used to exercise the generator's collision detection.
package other

// SameName has the same bare name as testdata.SameName.
type SameName struct {
	Value int
}
