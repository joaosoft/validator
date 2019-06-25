package validator

func (v *Validator) NewDefaultBeforeHandlers() map[string]BeforeTagHandler {
	return map[string]BeforeTagHandler{
		ConstTagId:   v.validate_id,
		ConstTagIf:   v.validate_if,
		ConstTagArgs: v.validate_args,
	}
}
