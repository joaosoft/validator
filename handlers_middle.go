package validator

func (v *Validator) newDefaultMiddleHandlers() map[string]middleTagHandler {
	return map[string]middleTagHandler{
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
		ConstTagString:   v.validate_string,
		ConstTagDistinct: v.validate_distinct,
		ConstTagKey:      v.validate_key,
		ConstTagAlpha:    v.validate_alpha,
		ConstTagNumeric:  v.validate_numeric,
		ConstTagBool:     v.validate_bool,
		ConstTagEncode:   v.validate_encode,
	}
}
