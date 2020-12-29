package validator

import "github.com/joaosoft/errors"

var (
	ErrorSkipValidation     = errors.New(errors.ErrorLevel, 1, "skip validation")
	ErrorInvalidValue       = errors.New(errors.ErrorLevel, 2, "invalid value")
	ErrorInvalidPointer     = errors.New(errors.ErrorLevel, 3, "invalid pointer")
	ErrorInvalidTag         = errors.New(errors.ErrorLevel, 4, "invalid tag [%s]")
	ErrorInvalidTagArgument = errors.New(errors.ErrorLevel, 5, "invalid tag argument [%s]")
	ErrorInvalidTagPrefix   = errors.New(errors.ErrorLevel, 6, "invalid prefix [%s] on tag [%s]")
)
