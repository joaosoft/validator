package validator

// Replace tags
const (
	constTagSplitValues    = ";"
	constTagReplaceStart   = "{{"
	constTagReplaceEnd     = "}}"
	constTagReplaceIdStart = "{"
	constTagReplaceIdEnd   = "}"
)

// Regexes
const (
	constRegexForReplaceId = "^" + constTagReplaceIdStart + "[A-Za-z0-9_-]+:?([A-Za-z0-9_-]+;?)+" + constTagReplaceIdEnd + "$"
	constRegexForReplace   = "^" + constTagReplaceStart + "[A-Za-z0-9_-]+:?([A-Za-z0-9_-]+;?)+" + constTagReplaceEnd + "$"
	constRegexForEmail     = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	constRegexForTrim      = "  +"
)

// Tags
const (
	constDefaultValidationTag = "validate"
	constDefaultLogTag        = "validator"

	constTagJson = "json"
)

// Validation prefix tags
const (
	constPrefixTagItem = "item"
	constPrefixTagKey  = "key"
)

// Validation tags
const (
	constTagId         = "id"
	constTagArg        = "arg"
	constTagValue      = "value"
	constTagError      = "error"
	constTagIf         = "if"
	constTagNot        = "not"
	constTagOptions    = "options"
	constTagNotOptions = "not-options"
	constTagSize       = "size"
	constTagMin        = "min"
	constTagMax        = "max"
	constTagNotEmpty   = "not-empty"
	constTagIsEmpty    = "is-empty"
	constTagNotNull    = "not-null"
	constTagIsNull     = "is-null"
	constTagRegex      = "regex"
	constTagCallback   = "callback"
	constTagAlpha      = "alpha"
	constTagNumeric    = "numeric"
	constTagBool       = "bool"
	constTagPassword   = "password"
	constTagArgs       = "args"
	constTagContains   = "contains"
	constTagPrefix     = "prefix"
	constTagSuffix     = "suffix"
	constTagUUID       = "uuid"
	constTagIp         = "ip"
	constTagIpV4       = "ipv4"
	constTagIpV6       = "ipv6"
	constTagBase64     = "base64"
	constTagEmail      = "email"
	constTagURL        = "url"
	constTagHex        = "hex"
	constTagFile       = "file"
)

// Validation set tags
const (
	constTagSet         = "set"
	constTagSetEmpty    = "set-empty"
	constTagSetDistinct = "set-distinct"
	constTagSetTrim     = "set-trim"
	constTagSetTitle    = "set-title"
	constTagSetLower    = "set-lower"
	constTagSetUpper    = "set-upper"
	constTagSetKey      = "set-key"
	constTagSetSanitize = "set-sanitize"
	constTagSetMd5      = "set-md5"
	constTagSetRandom   = "set-random"
)

// List of values
const (
	constAlphanumericLowerAlphabet = "abcdefghijklmnopqrstuvwxyzáéíóúãõâôàèìòùç"
	constAlphanumericUpperAlphabet = "ABCDEFGHUJKLMNOPQRSTUVWXYZÁÉÍÓÚÃÕÂÔÀÈÌÒÙÇ"
	constNumericAlphabet           = "0123456789"
	constSpecialAlphabet           = "!\"#$%&/()=?*@€£‰¶÷[]≠§±´`\\|~<>,;.:-_ "
)

// Condition tags
const (
	constConditionOk = "ok"
	constConditionKo = "ko"

	constParenthesesStart = "("
	constParenthesesEnd   = ")"
)

// Password checks
const (
	constPasswordCheckNumber      = "number"
	constPasswordCheckLetter      = "letter"
	constPasswordCheckSpace       = "space"
	constPasswordCheckUpper       = "upper"
	constPasswordCheckLower       = "lower"
	constPasswordCheckSymbol      = "symbol"
	constPasswordCheckPunctuation = "punctuation"
	constPasswordBlackListFile    = "./conf/password_black_list.txt"
)

// Minimum values
const (
	constMinNumeric     = 1
	constMinLetter      = 1
	constMinLower       = 1
	constMinUpper       = 1
	constMinSpace       = 0
	constMinSymbol      = 0
	constMinPunctuation = 1
	constMinLength      = 8
)
