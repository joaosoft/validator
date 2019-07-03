package validator

func (v *Validator) newDefaultPosHandlers() map[string]afterTagHandler {
	return map[string]afterTagHandler{
		ConstTagError: v.validate_error,
	}
}
