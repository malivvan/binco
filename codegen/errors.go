package codegen

import (
	"errors"
	"fmt"
)

// ErrCodeGenFailed is returned by Generate when the produced source does not compile.
var ErrCodeGenFailed = fmt.Errorf("code generation failed: ensure all codegen.Generator " +
	"options are set correctly; see the generated code for details")

// ErrNameCollision reports two types that would be generated under the same
// name. The second type needs an explicit name through genopts.WithSerializer.
var ErrNameCollision = errors.New("generated name collision")
