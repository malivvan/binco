package genopts

// Mode selects the codec family the generator emits.
type Mode int

const (
	ModeSafe Mode = iota
	ModeUnsafe
	ModeNotUnsafe
)
