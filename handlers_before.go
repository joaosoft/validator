package validator

func (v *Validator) newDefaultBeforeHandlers() map[string]beforeTagHandler {
	return map[string]beforeTagHandler{
		ConstTagId:   v.validate_id,
		ConstTagIf:   v.validate_if,
		ConstTagArgs: v.validate_args,
	}
}
