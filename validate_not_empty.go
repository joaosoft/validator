package validator

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	uuid "github.com/satori/go.uuid"
)

func (v *Validator) validate_is_null(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, _ := v._getValue(validationData.Value)
	if !isNil {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}

func (v *Validator) validate_not_null(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if errs := v.validate_is_null(context, validationData); len(errs) == 0 {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}

func (v *Validator) validate_is_empty(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	var isZero bool

	isNil, obj, value := v._getValue(validationData.Value)

	switch obj.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:

		switch obj.Type() {
		case reflect.TypeOf(uuid.UUID{}):
			if value.(uuid.UUID) == uuid.Nil {
				isZero = true
			}
		default:
			isZero = obj.Len() == 0
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		isZero = obj.Int() == 0
	case reflect.Float32, reflect.Float64:
		isZero = obj.Float() == 0
	case reflect.String:
		isZero = utf8.RuneCountInString(strings.TrimSpace(obj.String())) == 0
	case reflect.Bool:
		isZero = obj.Bool() == false
	case reflect.Struct:
		if reflect.DeepEqual(value, reflect.New(obj.Type()).Interface()) {
			isZero = true
		}
	default:
		if isNil {
			isZero = true
		}
	}

	if !isZero {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}

func (v *Validator) validate_regex(context *ValidatorContext, validationData *ValidationData) []error {

	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	r, err := regexp.Compile(validationData.Expected.(string))
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if utf8.RuneCountInString(strValue) > 0 {
		if !r.MatchString(strValue) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_callback(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	validators := strings.Split(validationData.Expected.(string), constTagSplitValues)

	for _, validator := range validators {
		if callback, ok := v.callbacks[validator]; ok {
			errs := callback(context, validationData)
			if errs != nil && len(errs) > 0 {
				rtnErrs = append(rtnErrs, errs...)
			}

			if !v.canValidateAll {
				return rtnErrs
			}
		}
	}

	return rtnErrs
}

type ErrorValidate struct {
	error
	replaced bool
}

func (v *Validator) validate_alpha(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return rtnErrs
	}

	for _, r := range strValue {
		if !unicode.IsLetter(r) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
			break
		}
	}

	return rtnErrs
}

func (v *Validator) validate_numeric(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return rtnErrs
	}

	for _, r := range strValue {
		if !unicode.IsNumber(r) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
			break
		}
	}

	return rtnErrs
}

func (v *Validator) validate_bool(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return rtnErrs
	}

	switch strings.ToLower(strValue) {
	case "true", "false":
	default:
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}

func (v *Validator) validate_password(context *ValidatorContext, validationData *ValidationData) (errs []error) {
	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return nil
	}

	return v.password.settings.Compare(strValue)
}

func (v *Validator) validate_prefix(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if !strings.HasPrefix(v._convertToString(value), v._convertToString(expected)) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_suffix(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if !strings.HasSuffix(v._convertToString(value), v._convertToString(expected)) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_contains(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if !strings.Contains(v._convertToString(value), v._convertToString(expected)) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_uuid(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)
	check := false

	_, obj, value := v._getValue(validationData.Value)

	var checkValue interface{}
	switch obj.Type() {
	case reflect.TypeOf(uuid.UUID{}):
		check = true
		checkValue = obj.Interface().(uuid.UUID).String()
	case reflect.TypeOf(""):
		check = true
		checkValue = value
	}

	if check {
		if _, err := uuid.FromString(v._convertToString(checkValue)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_ip(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if ip := net.ParseIP(v._convertToString(value)); ip == nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_ipv4(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if ip := net.ParseIP(v._convertToString(value)); ip == nil || ip.To4() == nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_ipv6(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if ip := net.ParseIP(v._convertToString(value)); ip == nil || ip.To16() == nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_email(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		r, err := regexp.Compile(constRegexForEmail)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if !r.MatchString(v._convertToString(value)) {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_url(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if _, err := url.ParseRequestURI(v._convertToString(value)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_base64(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if _, err := base64.StdEncoding.DecodeString(v._convertToString(value)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_hex(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if _, err := hex.DecodeString(v._convertToString(value)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_file(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, _, value := v._getValue(validationData.Value)

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		if _, err := os.Stat(v._convertToString(value)); err != nil {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()
	if err = _setValue(kind, obj, expected); err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	return rtnErrs
}

func (v *Validator) validate_set_empty(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	obj.Set(reflect.Zero(reflect.TypeOf(value)))

	return rtnErrs
}

func (v *Validator) validate_set_trim(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.TrimSpace(value.(string))

		r, err := regexp.Compile(constRegexForTrim)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		newValue = string(r.ReplaceAll(bytes.TrimSpace([]byte(newValue)), []byte(" ")))
		if err = _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set_title(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.Title(value.(string))
		if err := _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set_upper(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.ToUpper(value.(string))
		if err := _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set_lower(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.ToLower(value.(string))
		if err := _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set_md5(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if expected == "" {
			expected = value
		}

		newValue := fmt.Sprintf("%x", md5.Sum([]byte(v._convertToString(expected))))
		if err = _setValue(kind, obj, newValue); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set_key(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if expected == "" {
			expected = value
		}

		if err = _setValue(kind, obj, convertToKey(strings.TrimSpace(v._convertToString(expected)), true)); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set_random(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		rtnErrs = append(rtnErrs, ErrorInvalidPointer)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()
	isPointer := false

	if kind == reflect.Ptr && !obj.IsNil() {
		isPointer = true
		value = obj.Elem()
		kind = reflect.TypeOf(obj.Interface()).Kind()
	}

	switch kind {
	case reflect.String:
		expected, err := v._loadExpectedValue(context, validationData.Expected)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if v._convertToString(expected) == "" {
			expected = value
		}

		if err = _setValue(kind, obj, v._random(v._convertToString(expected))); err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min := 1
		max := 100
		obj.SetInt(int64(rand.Intn(max-min) + min))

	case reflect.Float32:
		obj.SetFloat(float64(rand.Float32()))

	case reflect.Float64:
		obj.SetFloat(rand.Float64())

	case reflect.Bool:
		obj.SetBool(rand.Intn(1) == 1)
	}

	if isPointer {
		value = obj.Addr()
	}

	return rtnErrs
}

func (v *Validator) validate_set_distinct(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, parentObj, parentValue := v._getValue(validationData.Parent)

	if parentObj.CanAddr() {
		kind := reflect.TypeOf(parentValue).Kind()

		if kind != reflect.Array && kind != reflect.Slice {
			return rtnErrs
		}
		newInstance := reflect.New(parentObj.Type()).Elem()

		values := make(map[interface{}]bool)
		for i := 0; i < parentObj.Len(); i++ {

			indexValue := parentObj.Index(i)
			if indexValue.Kind() == reflect.Ptr && !indexValue.IsNil() {
				indexValue = parentObj.Index(i).Elem()
			}

			if _, ok := values[indexValue.Interface()]; ok {
				continue
			}

			switch indexValue.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Float32, reflect.Float64,
				reflect.String,
				reflect.Bool:
				if parentObj.Index(i).Kind() == reflect.Ptr && !parentObj.Index(i).IsNil() {
					newInstance = reflect.Append(newInstance, indexValue.Addr())
				} else {
					newInstance = reflect.Append(newInstance, indexValue)
				}

				values[indexValue.Interface()] = true
			}
		}

		// set the new instance without duplicated values
		parentObj.Set(newInstance)
	}

	return rtnErrs
}

func (v *Validator) validate_set_sanitize(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	split := strings.Split(validationData.Expected.(string), constTagSplitValues)
	invalid := make([]string, 0)

	// validate expected
	for _, str := range split {
		if strings.Contains(strValue, str) {
			invalid = append(invalid, str)
		}
	}

	// validate global
	for _, str := range v.sanitize {
		if strings.Contains(strValue, str) {
			invalid = append(invalid, str)
		}
	}

	if len(invalid) > 0 {
		rtnErrs = append(rtnErrs, ErrorInvalidValue)
	}

	return rtnErrs
}
