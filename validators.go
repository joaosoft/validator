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
	"strconv"
	"strings"
	"time"
	"unicode"

	"errors"

	"github.com/satori/go.uuid"
)

func (v *Validator) validate_value(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if strValue != v._convertToString(expected) {
		err := fmt.Errorf("the value [%+v] is different of the expected [%+v] on field [%s]", value, expected, validationData.Name)
		rtnErrs = append(rtnErrs, err)
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

	split := strings.Split(validationData.Expected.(string), ";")
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
		err := fmt.Errorf("the value [%+v] is has invalid characters [%+v] on field [%s]", value, strings.Join(invalid, ","), validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_not(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if strValue == v._convertToString(expected) {
		err := fmt.Errorf("the expected [%+v] should be different of the [%+v] on field [%s]", expected, value, validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_options(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, obj, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	options := strings.Split(validationData.Expected.(string), ";")
	var invalidValue interface{}

	switch obj.Kind() {
	case reflect.Array, reflect.Slice:
		var err error
		var opt interface{}
		optionsVal := make(map[string]bool)
		for _, option := range options {
			opt, err = v._loadExpectedValue(context, option)
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					return rtnErrs
				} else {
					continue
				}
			}
			optionsVal[v._convertToString(opt)] = true
		}

		for i := 0; i < obj.Len(); i++ {
			nextValue := obj.Index(i)

			if !nextValue.CanInterface() {
				continue
			}

			_, ok := optionsVal[v._convertToString(nextValue.Interface())]
			if !ok {
				invalidValue = nextValue.Interface()
				err := fmt.Errorf("the value [%+v] is different of the expected options [%+v] on field [%s]", invalidValue, validationData.Expected, validationData.Name)
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					break
				}
			}
		}

	case reflect.Map:
		optionsMap := make(map[string]interface{})
		var value interface{}
		for _, option := range options {
			values := strings.Split(option, ":")
			if len(values) != 2 {
				continue
			}

			var err error
			value, err = v._loadExpectedValue(context, values[1])
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					return rtnErrs
				} else {
					continue
				}
			}

			optionsMap[values[0]] = value
		}

		for _, key := range obj.MapKeys() {
			nextValue := obj.MapIndex(key)

			if !nextValue.CanInterface() {
				continue
			}

			val, ok := optionsMap[v._convertToString(key.Interface())]
			if !ok || v._convertToString(nextValue.Interface()) != v._convertToString(val) {
				invalidValue = fmt.Sprintf("%s:%s", v._convertToString(key.Interface()), v._convertToString(nextValue.Interface()))
				err := fmt.Errorf("the value [%+v] is different of the expected options [%+v] on field [%s]", nextValue.Interface(), validationData.Expected, validationData.Name)
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					break
				}
			}
		}

	default:
		var err error
		var opt interface{}
		optionsVal := make(map[string]bool)
		for _, option := range options {
			opt, err = v._loadExpectedValue(context, option)
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					return rtnErrs
				} else {
					continue
				}
			}
			optionsVal[v._convertToString(opt)] = true
		}

		_, ok := optionsVal[v._convertToString(value)]
		if !ok {
			invalidValue = value
			err := fmt.Errorf("the value [%+v] is different of the expected options [%+v] on field [%s]", invalidValue, validationData.Expected, validationData.Name)
			rtnErrs = append(rtnErrs, err)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_size(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, obj, value := v._getValue(validationData.Value)
	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	size, e := strconv.Atoi(v._convertToString(expected))
	if e != nil {
		err := fmt.Errorf("the size [%s] is invalid on field [%s] value [%+v]", expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	var valueSize int64

	switch obj.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		valueSize = int64(obj.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueSize = int64(len(strings.TrimSpace(strconv.Itoa(int(obj.Int())))))
	case reflect.Float32, reflect.Float64:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatFloat(obj.Float(), 'g', 1, 64))))
	case reflect.String:
		valueSize = int64(len(strings.TrimSpace(obj.String())))
	case reflect.Bool:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatBool(obj.Bool()))))
	default:
		if isNil {
			break
		}
		valueSize = int64(len(strings.TrimSpace(obj.String())))
	}

	if valueSize != int64(size) {
		err := fmt.Errorf("the length [%+v] is lower then the expected [%+v] on field [%s] value [%+v]", valueSize, expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_min(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	isNil, obj, value := v._getValue(validationData.Value)
	min, e := strconv.Atoi(v._convertToString(expected))
	if e != nil {
		err := fmt.Errorf("the size [%+v] is invalid on field [%s] value [%+v]", expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	var valueSize int64

	switch obj.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		valueSize = int64(obj.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueSize = obj.Int()
	case reflect.Float32, reflect.Float64:
		valueSize = int64(obj.Float())
	case reflect.String:
		valueSize = int64(len(strings.TrimSpace(obj.String())))
	case reflect.Bool:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatBool(obj.Bool()))))
	default:
		if isNil {
			break
		}
		valueSize = int64(len(strings.TrimSpace(obj.String())))
	}

	if valueSize < int64(min) {
		err := fmt.Errorf("the length [%+v] is lower then the expected [%+v] on field [%s] value [%+v]", valueSize, expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_max(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	expected, err := v._loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	isNil, obj, value := v._getValue(validationData.Value)
	max, e := strconv.Atoi(v._convertToString(expected))
	if e != nil {
		err := fmt.Errorf("the size [%s+v is invalid on field [%s] value [%+v]", validationData.Expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	var valueSize int64

	switch obj.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		valueSize = int64(obj.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueSize = obj.Int()
	case reflect.Float32, reflect.Float64:
		valueSize = int64(obj.Float())
	case reflect.String:
		valueSize = int64(len(strings.TrimSpace(obj.String())))
	case reflect.Bool:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatBool(obj.Bool()))))
	default:
		if isNil {
			break
		}
		valueSize = int64(len(strings.TrimSpace(obj.String())))
	}

	if valueSize > int64(max) {
		err := fmt.Errorf("the length [%+v] is bigger then the expected [%+v] on field [%s] value [%+v]", valueSize, expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_notzero(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if errs := v.validate_iszero(context, validationData); len(errs) == 0 {
		err := fmt.Errorf("the value shouldn't be zero on field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_isnull(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	if !isNil {
		err := fmt.Errorf("the value should be null on field [%s] instead of [%+v]", validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_notnull(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if errs := v.validate_isnull(context, validationData); len(errs) == 0 {
		err := fmt.Errorf("the value shouldn't be null on field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_iszero(context *ValidatorContext, validationData *ValidationData) []error {
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
		isZero = len(strings.TrimSpace(obj.String())) == 0
	case reflect.Bool:
		isZero = obj.Bool() == false
	case reflect.Struct:
		if value == reflect.Zero(obj.Type()).Interface() {
			isZero = true
		}
	default:
		if isNil {
			isZero = true
		}
	}

	if !isZero {
		err := fmt.Errorf("the value should be zero on field [%s] instead of [%+v]", validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
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

	if len(strValue) > 0 {
		if !r.MatchString(strValue) {
			err := fmt.Errorf("invalid value [%s] on field [%s]", strValue, validationData.Name)
			rtnErrs = append(rtnErrs, err)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_callback(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	validators := strings.Split(validationData.Expected.(string), ";")

	for _, validator := range validators {
		if callback, ok := v.callbacks[validator]; ok {
			errs := callback(context, validationData)
			if errs != nil && len(errs) > 0 {
				rtnErrs = append(rtnErrs, errs...)
			}

			if !v.validateAll {
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

func (v *Validator) validate_error(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)
	added := make(map[string]bool)
	for i, e := range *validationData.Errors {
		if _, ok := validationData.ErrorsReplaced[e]; ok {
			continue
		}
		if v.errorCodeHandler != nil {
			var expected string

			if validationData.Expected != nil {
				expected = validationData.Expected.(string)
			}

			if matched, err := regexp.MatchString(ConstRegexForTagValue, expected); err != nil {
				rtnErrs = append(rtnErrs, err)
			} else {
				if matched {
					replacer := strings.NewReplacer("{", "", "}", "")
					expected := replacer.Replace(validationData.Expected.(string))

					split := strings.SplitN(expected, ":", 2)
					if len(split) == 0 {
						rtnErrs = append(rtnErrs, fmt.Errorf("invalid tag error defined [%s]", expected))
						continue
					}

					if _, ok := added[split[0]]; !ok {
						var arguments []interface{}
						if len(split) == 2 {
							splitArgs := strings.Split(split[1], ";")
							for _, arg := range splitArgs {
								arguments = append(arguments, arg)
							}
						}

						validationData.ErrorData = &ErrorData{
							Code:      split[0],
							Arguments: arguments,
						}

						newErr := v.errorCodeHandler(context, validationData)
						if newErr != nil {
							(*validationData.Errors)[i] = newErr
							validationData.ErrorsReplaced[newErr] = true
						}

						added[split[0]] = true
					} else {
						if len(*validationData.Errors)-1 == i {
							*validationData.Errors = (*validationData.Errors)[:i]
						} else {
							*validationData.Errors = append((*validationData.Errors)[:i], (*validationData.Errors)[i+1:]...)
						}
					}
				} else {
					newErr := errors.New(expected)
					(*validationData.Errors)[i] = newErr
					validationData.ErrorsReplaced[newErr] = true
				}
			}
		}
	}

	return rtnErrs
}

func (v *Validator) validate_id(context *ValidatorContext, validationData *ValidationData) []error {
	return nil
}

func (v *Validator) validate_if(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	str := validationData.Expected.(string)
	var expressions []*expression
	var expr *expression
	var query string

	// read conditions
	size := len(str)

	for i := 0; i < size; i++ {
		switch str[i] {
		case '(':
			continue

		case ')':
			startId := strings.Index(query, fmt.Sprintf("%s=", ConstTagId))
			startArg := strings.Index(query, fmt.Sprintf("%s=", ConstTagArg))
			if startId == -1 && startArg == -1 {
				return rtnErrs
			}

			var start int
			var tag string
			if startId > -1 {
				tag = ConstTagId
				start = startId
			}

			if startArg > -1 {
				tag = ConstTagArg
				start = startArg
			}

			end := strings.Index(query[start:], " ")
			if end == -1 {
				end = size - 1
			}

			id := query[start+len(tag)+1 : end]
			query = query[end+1:]

			data, ok := context.GetValue(tag, id)

			if ok {
				var errs []error
				err := context.execute(data.typ, data.value, strings.Split(query, " "), &errs)

				// get next operator
				var operator Operator
				if index := strings.Index(str[i+1:], "("); index > -1 {
					operator = Operator(strings.TrimSpace(str[i+1 : i+1+index]))

					str = str[i+1+index:]
					i = 0
					size = len(str)
				}

				expr = &expression{
					data:         data,
					result:       err,
					nextOperator: operator,
					expected:     query,
				}
				expressions = append(expressions, expr)
			}

			query = ""

		default:
			query = fmt.Sprintf("%s%c", query, str[i])
		}
	}

	// validate all conditions
	var condition = ""
	var prevOperator = NONE

	for _, expr := range expressions {

		if condition == "" {
			if expr.result == nil {
				condition = "ok"
			} else {
				condition = "ko"
			}
		} else {

			switch prevOperator {
			case AND:
				if expr.result != nil {
					condition = "ko"
				}
			case OR:
				if expr.result == nil && condition == "ko" {
					condition = "ok"
				}
			case NONE:
				if expr.result == nil {
					condition = "ok"
				}
			}
		}

		prevOperator = expr.nextOperator
	}

	if condition == "ko" {
		return []error{ErrorSkipValidation}
	}

	return nil
}

func (v *Validator) validate_set_trim(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.TrimSpace(value.(string))

		r, err := regexp.Compile("  +")
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		newValue = string(r.ReplaceAll(bytes.TrimSpace([]byte(newValue)), []byte(" ")))
		setValue(kind, obj, newValue)
	}

	return rtnErrs
}

func (v *Validator) validate_set_title(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.Title(value.(string))
		setValue(kind, obj, newValue)
	}

	return rtnErrs
}

func (v *Validator) validate_set_upper(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.ToUpper(value.(string))
		setValue(kind, obj, newValue)
	}

	return rtnErrs
}

func (v *Validator) validate_set_lower(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	kind := reflect.TypeOf(value).Kind()

	switch kind {
	case reflect.String:
		newValue := strings.ToLower(value.(string))
		setValue(kind, obj, newValue)
	}

	return rtnErrs
}

func (v *Validator) validate_set(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	newExpected := v._convertToString(validationData.Expected)
	if matched, err := regexp.MatchString(ConstRegexForTagValue, newExpected); err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	} else {
		if matched {
			replacer := strings.NewReplacer("{", "", "}", "")
			id := replacer.Replace(newExpected)
			validationData.Expected = value

			if newValue, ok := context.GetValue(ConstTagId, id); ok {
				value := obj.FieldByName(validationData.Field)
				kind := reflect.TypeOf(value).Kind()

				setValue(kind, value, newValue.value.Interface())
			} else {
				err := fmt.Errorf("invalid set tag [%+v] on field [%s]", validationData.Expected, validationData.Name)
				rtnErrs = append(rtnErrs, err)
				return rtnErrs
			}
		} else {
			kind := reflect.TypeOf(value).Kind()
			setValue(kind, obj, validationData.Expected)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_set_md5(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
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
		setValue(kind, obj, newValue)
	}

	return rtnErrs
}

func (v *Validator) validate_set_key(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
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

		setValue(kind, obj, convertToKey(strings.TrimSpace(v._convertToString(expected)), true))
	}

	return rtnErrs
}

func (v *Validator) validate_set_random(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v._getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
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

		if v._convertToString(expected) == "" {
			expected = value
		}

		setValue(kind, obj, v._random(v._convertToString(expected)))
	}

	return rtnErrs
}

func setValue(kind reflect.Kind, obj reflect.Value, newValue interface{}) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, _ := strconv.Atoi(newValue.(string))
		obj.SetInt(int64(v))
	case reflect.Float32, reflect.Float64:
		v, _ := strconv.ParseFloat(newValue.(string), 64)
		obj.SetFloat(v)
	case reflect.String:
		obj.SetString(newValue.(string))
	case reflect.Bool:
		v, _ := strconv.ParseBool(newValue.(string))
		obj.SetBool(v)
	}
}

func (v *Validator) validate_distinct(context *ValidatorContext, validationData *ValidationData) []error {
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

func (v *Validator) validate_alpha(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)

	if strValue == "" || isNil {
		return rtnErrs
	}

	for _, r := range strValue {
		if !unicode.IsLetter(r) {
			err := fmt.Errorf("the value [%+v] is invalid for type alphanumeric on field [%s] value [%+v]", value, validationData.Name, value)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] is invalid for type numeric on field [%s] value [%+v]", value, validationData.Name, value)
			rtnErrs = append(rtnErrs, err)
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
		err := fmt.Errorf("the value [%+v] is invalid for type bool on field [%s] value [%+v]", value, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
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
			err := fmt.Errorf("the value on field [%s] should have the prefix [%+v] instead of [%+v]", validationData.Name, expected, value)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value on field [%s] should have the suffix to [%+v] instead of [%+v]", validationData.Name, expected, value)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value on field [%s] should contain [%+v] instead of [%+v]", validationData.Name, expected, value)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%s] on field [%s] should be a valid UUID", checkValue, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid IP", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid IPv4", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid IPv6", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
		r, err := regexp.Compile(ConstRegexForEmail)
		if err != nil {
			rtnErrs = append(rtnErrs, err)
			return rtnErrs
		}

		if !r.MatchString(v._convertToString(value)) {
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid Email", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid URL", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid Base64", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid Hexadecimal", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
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
			err := fmt.Errorf("the value [%+v] on field [%s] should be a valid File", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_args(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	splitArgs := strings.Split(validationData.Expected.(string), ";")

	for _, arg := range splitArgs {
		validationData.Arguments = append(validationData.Arguments, arg)
	}

	return rtnErrs
}

func (v *Validator) _convertToString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%+v", value)
}

func (v *Validator) _getValue(value reflect.Value) (isNil bool, _ reflect.Value, _ interface{}) {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return true, value, value.Interface()
		}
		return value.Elem().Interface() == nil, value.Elem(), value.Elem().Interface()
	}

	return value.Interface() == nil, value, value.Interface()
}

func (v *Validator) _loadExpectedValue(context *ValidatorContext, expected interface{}) (interface{}, error) {

	if expected != nil && v._convertToString(expected) != "" {
		strValue := v._convertToString(expected)
		if matched, err := regexp.MatchString(ConstRegexForTagValue, strValue); err != nil {
			return "", err
		} else {
			if matched {
				replacer := strings.NewReplacer("{", "", "}", "")
				id := replacer.Replace(strValue)

				value, ok := context.GetValue(ConstTagId, id)
				if !ok {
					value, ok = context.GetValue(ConstTagArg, id)
					if !ok {
						value, ok = context.GetValue(ConstTagJson, id)
					}
				}

				if ok {
					return value.value.Interface(), nil
				}
			}
		}

	}
	return expected, nil
}

func (v *Validator) _random(strValue string) string {
	rand.Seed(time.Now().UnixNano())
	alphabetLowerChars := []rune(ConstAlphanumericLowerAlphabet)
	alphabetUpperChars := []rune(ConstAlphanumericUpperAlphabet)
	alphabetNumbers := []rune(ConstNumericAlphabet)
	alphabetSpecial := []rune(ConstSpecialAlphabet)

	newValue := []rune(strValue)

	for i, char := range newValue {
		if !unicode.IsSpace(char) {
			var alphabet []rune
			if unicode.IsLetter(char) {
				if unicode.IsUpper(char) {
					alphabet = alphabetUpperChars
				} else {
					alphabet = alphabetLowerChars
				}
			} else if unicode.IsNumber(char) {
				alphabet = alphabetNumbers
			} else {
				alphabet = alphabetSpecial
			}

			newValue[i] = alphabet[rand.Intn(len(alphabet))]
		}
	}

	return string(newValue)
}
