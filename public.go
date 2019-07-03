package validator

func AddBefore(name string, handler beforeTagHandler) *Validator {
	return validatorInstance.AddBefore(name, handler)
}

func AddMiddle(name string, handler middleTagHandler) *Validator {
	return validatorInstance.AddMiddle(name, handler)
}

func AddAfter(name string, handler afterTagHandler) *Validator {
	return validatorInstance.AddAfter(name, handler)
}

func SetValidateAll(validate bool) *Validator {
	return validatorInstance.SetValidateAll(validate)
}

func SetTag(tag string) *Validator {
	return validatorInstance.SetTag(tag)
}

func SetSanitize(sanitize []string) *Validator {
	return validatorInstance.SetSanitize(sanitize)
}

// Validate ...
func Validate(obj interface{}, args ...*Argument) []error {
	return validatorInstance.Validate(obj, args...)
}
