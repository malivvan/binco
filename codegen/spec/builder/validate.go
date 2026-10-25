package builder

import (
	"fmt"
	"go/parser"
	"go/token"
	"strconv"
)

// ValidateExpr checks that expr is a single Go expression.
//
// Validators and migrations are copied into generated source verbatim, so a
// typo used to surface only at the final format step as a generic "code
// generation failed". Checking here names the offending option instead.
func ValidateExpr(option, expr string) error {
	if expr == "" {
		return nil
	}
	if _, err := parser.ParseExpr(expr); err != nil {
		return fmt.Errorf("%s %s is not a Go expression: %w",
			option, strconv.Quote(expr), err)
	}
	return nil
}

// ValidateIdent checks that name is a single Go identifier.
func ValidateIdent(option, name string) error {
	if name == "" {
		return nil
	}
	if !token.IsIdentifier(name) {
		return fmt.Errorf("%s %s is not a Go identifier", option, strconv.Quote(name))
	}
	return nil
}
