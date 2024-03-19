package validator

import (
	"strings"
)

func (v *Validator) validate_args(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	splitArgs := strings.Split(validationData.Expected.(string), constTagSplitValues)

	for _, arg := range splitArgs {
		validationData.Arguments = append(validationData.Arguments, arg)
	}

	return rtnErrs
}
