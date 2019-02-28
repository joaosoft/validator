package validator

func (v *Validator) NewDefaultMiddleHandlers() map[string]MiddleTagHandler {
	return map[string]MiddleTagHandler{
		ConstTagValue:    v.validate_value,
		ConstTagNot:      v.validate_not,
		ConstTagOptions:  v.validate_options,
		ConstTagSize:     v.validate_size,
		ConstTagMin:      v.validate_min,
		ConstTagMax:      v.validate_max,
		ConstTagNotZero:  v.validate_notzero,
		ConstTagIsZero:   v.validate_iszero,
		ConstTagNotNull:  v.validate_notnull,
		ConstTagIsNull:   v.validate_isnull,
		ConstTagRegex:    v.validate_regex,
		ConstTagSpecial:  v.validate_special,
		ConstTagSanitize: v.validate_sanitize,
		ConstTagCallback: v.validate_callback,
		ConstTagSet:      v.validate_set,
		ConstTagDistinct: v.validate_distinct,
		ConstTagKey:      v.set_key,
		ConstTagAlpha:    v.validate_alpha,
		ConstTagNumeric:  v.validate_numeric,
		ConstTagBool:     v.validate_bool,
	}
}
