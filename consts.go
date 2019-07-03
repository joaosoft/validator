package validator

const (
	ConstSpecialTagForDateDefault  = "date"
	ConstSpecialTagForDateYYYYMMDD = "YYYYMMDD"
	ConstSpecialTagForDateDDMMYYYY = "DDMMYYYY"
	ConstSpecialTagForTimeDefault  = "time"
	ConstSpecialTagForTimeHHMMSS   = "HHMMSS"
	ConstSpecialTagForURL          = "url"
	ConstSpecialTagForEmail        = "email"

	ConstSetTagForTrim  = "trim"
	ConstSetTagForTitle = "title"
	ConstSetTagForUpper = "upper"
	ConstSetTagForLower = "lower"
	ConstSetTagForKey   = "key"

	ConstEncodeMd5    = "md5"
	ConstEncodeRandom = "random"

	ConstRegexForDateDDMMYYYY = `^(0?[1-9]|[12][0-9]|3[01])(/|-|.)([1-9]|0[0-9]|1[0-2])(/|-|.)[0-9]{4}$`
	ConstRegexForDateYYYYMMDD = `^[0-9]{4}(/|-|.)(0?[1-9]|[12][0-9]|3[01])(/|-|.)([1-9]|0[0-9]|1[0-2])$`
	ConstRegexForDateDefault  = ConstRegexForDateDDMMYYYY
	ConstRegexForTimeDefault  = ConstRegexForTimeHHMMSS
	ConstRegexForTimeHHMMSS   = `^([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$`
	ConstRegexForTagValue     = "{[A-Za-z0-9_-]+:?([A-Za-z0-9_-];?)+}"
	ConstRegexForURL          = "^((http|https)://)?(www)?[a-zA-Z0-9-._:/?&=,]+$"
	ConstRegexForEmail        = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

	ConstDefaultValidationTag = "validate"
	ConstDefaultLogTag        = "validator"

	ConstPrefixTagItem = "item"
	ConstPrefixTagKey  = "key"

	ConstTagId       = "id"
	ConstTagArg      = "arg"
	ConstTagValue    = "value"
	ConstTagError    = "error"
	ConstTagIf       = "if"
	ConstTagNot      = "not"
	ConstTagOptions  = "options"
	ConstTagSize     = "size"
	ConstTagMin      = "min"
	ConstTagMax      = "max"
	ConstTagNotZero  = "notzero"
	ConstTagIsZero   = "iszero"
	ConstTagNotNull  = "notnull"
	ConstTagIsNull   = "isnull"
	ConstTagRegex    = "regex"
	ConstTagSpecial  = "special"
	ConstTagSanitize = "sanitize"
	ConstTagCallback = "callback"
	ConstTagSet      = "set"
	ConstTagString   = "string"
	ConstTagKey      = "key"
	ConstTagDistinct = "distinct"
	ConstTagAlpha    = "alpha"
	ConstTagNumeric  = "numeric"
	ConstTagBool     = "bool"
	ConstTagDecode   = "decode"
	ConstTagEncode   = "encode"
	ConstTagArgs     = "args"
)
