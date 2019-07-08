package validator

const (
	ConstRegexForTagValue = "{[A-Za-z0-9_-]+:?([A-Za-z0-9_-];?)+}"
	ConstRegexForEmail    = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

	ConstDefaultValidationTag = "validate"
	ConstDefaultLogTag        = "validator"

	ConstPrefixTagItem = "item"
	ConstPrefixTagKey  = "key"

	ConstTagJson = "json"

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
	ConstTagCallback = "callback"
	ConstTagAlpha    = "alpha"
	ConstTagNumeric  = "numeric"
	ConstTagBool     = "bool"
	ConstTagArgs     = "args"
	ConstTagContains = "contains"
	ConstTagPrefix   = "prefix"
	ConstTagSuffix   = "suffix"
	ConstTagUUID     = "uuid"
	ConstTagIp       = "ip"
	ConstTagIpV4     = "ipv4"
	ConstTagIpV6     = "ipv6"
	ConstTagBase64   = "base64"
	ConstTagEmail    = "email"
	ConstTagURL      = "url"
	ConstTagHex      = "hex"
	ConstTagFile     = "file"

	ConstTagSet         = "set"
	ConstTagSetDistinct = "set-distinct"
	ConstTagSetTrim     = "set-trim"
	ConstTagSetTitle    = "set-title"
	ConstTagSetLower    = "set-lower"
	ConstTagSetUpper    = "set-upper"
	ConstTagSetKey      = "set-key"
	ConstTagSetSanitize = "set-sanitize"
	ConstTagSetMd5      = "set-md5"
	ConstTagSetRandom   = "set-random"

	ConstAlphanumericAlphabet = "abcdefghijklmnopqrstuvwxyz"
	ConstNumericAlphabet      = "0123456789"
)
