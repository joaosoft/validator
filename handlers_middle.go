package validator

func (v *Validator) newDefaultMiddleHandlers() map[string]middleTagHandler {
	return map[string]middleTagHandler{
		ConstTagValue:       v.validate_value,
		ConstTagNot:         v.validate_not,
		ConstTagOptions:     v.validate_options,
		ConstTagNotOptions:  v.validate_notoptions,
		ConstTagSize:        v.validate_size,
		ConstTagMin:         v.validate_min,
		ConstTagMax:         v.validate_max,
		ConstTagNotZero:     v.validate_notzero,
		ConstTagIsZero:      v.validate_iszero,
		ConstTagNotNull:     v.validate_notnull,
		ConstTagIsNull:      v.validate_isnull,
		ConstTagRegex:       v.validate_regex,
		ConstTagCallback:    v.validate_callback,
		ConstTagSetDistinct: v.validate_distinct,
		ConstTagAlpha:       v.validate_alpha,
		ConstTagNumeric:     v.validate_numeric,
		ConstTagBool:        v.validate_bool,
		ConstTagPrefix:      v.validate_prefix,
		ConstTagSuffix:      v.validate_suffix,
		ConstTagContains:    v.validate_contains,
		ConstTagUUID:        v.validate_uuid,
		ConstTagIp:          v.validate_ip,
		ConstTagIpV4:        v.validate_ipv4,
		ConstTagIpV6:        v.validate_ipv6,
		ConstTagBase64:      v.validate_base64,
		ConstTagEmail:       v.validate_email,
		ConstTagURL:         v.validate_url,
		ConstTagHex:         v.validate_hex,
		ConstTagFile:        v.validate_file,

		ConstTagSet:         v.validate_set,
		ConstTagSetTrim:     v.validate_set_trim,
		ConstTagSetTitle:    v.validate_set_title,
		ConstTagSetLower:    v.validate_set_lower,
		ConstTagSetUpper:    v.validate_set_upper,
		ConstTagSetKey:      v.validate_set_key,
		ConstTagSetSanitize: v.validate_set_sanitize,
		ConstTagSetMd5:      v.validate_set_md5,
		ConstTagSetRandom:   v.validate_set_random,
	}
}
