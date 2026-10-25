package codegen

import (
	"fmt"
)

// ErrCodeGenFailed is returned by Generate when the produced source does not compile.
var ErrCodeGenFailed = fmt.Errorf("code generation failed: ensure all codegen.Generator " +
	"options are set correctly; see the generated code for details")
