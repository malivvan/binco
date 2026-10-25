package scanner

import (
	"fmt"
)

// NewUnsupportedQualifiedNameError reports a name the scanner cannot parse.
func NewUnsupportedQualifiedNameError[T QualifiedName](name T) error {
	return fmt.Errorf("unsupported '%v' QualifiedTypeName", name)
}
