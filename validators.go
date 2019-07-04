package validator

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"errors"

	"github.com/satori/go.uuid"
)

func (v *Validator) getValue(value reflect.Value) (isNull bool, _ reflect.Value, _ interface{}) {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return true, value, value.Interface()
		}
		return value.Elem().Interface() == nil, value.Elem(), value.Elem().Interface()
	}

	return value.Interface() == nil, value, value.Interface()
}

func (v *Validator) loadExpectedValue(context *ValidatorContext, expected interface{}) (string, error) {
	var newExpected string

	if expected != nil {
		newExpected = fmt.Sprintf("%+v", expected)
	}

	if matched, err := regexp.MatchString(ConstRegexForTagValue, newExpected); err != nil {
		return "", err
	} else {
		if matched {
			replacer := strings.NewReplacer("{", "", "}", "")
			id := replacer.Replace(newExpected)

			value, ok := context.GetValue(ConstTagId, id)
			if !ok {
				value, ok = context.GetValue(ConstTagArg, id)
			}

			if ok {
				newExpected = fmt.Sprintf("%+v", value.obj.Interface())
			}
		}
	}

	return newExpected, nil
}

func (v *Validator) validate_value(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNull, _, value := v.getValue(validationData.Value)
	if isNull || fmt.Sprintf("%+v", value) == "" {
		return rtnErrs
	}

	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if fmt.Sprintf("%+v", value) != expected {
		err := fmt.Errorf("the value [%+v] is different of the expected [%+v] on field [%s]", value, expected, validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_sanitize(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNull, _, value := v.getValue(validationData.Value)
	if isNull || fmt.Sprintf("%+v", value) == "" {
		return rtnErrs
	}

	val := fmt.Sprintf("%+v", value)
	split := strings.Split(validationData.Expected.(string), ";")
	invalid := make([]string, 0)

	// validate expected
	for _, str := range split {
		if strings.Contains(val, str) {
			invalid = append(invalid, str)
		}
	}

	// validate global
	for _, str := range v.sanitize {
		if strings.Contains(val, str) {
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

	isNull, _, value := v.getValue(validationData.Value)
	if isNull || fmt.Sprintf("%+v", value) == "" {
		return rtnErrs
	}

	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if fmt.Sprintf("%+v", value) == fmt.Sprintf("%+v", expected) {
		err := fmt.Errorf("the expected [%+v] should be different of the [%+v] on field [%s]", expected, value, validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_options(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNull, obj, value := v.getValue(validationData.Value)
	if isNull || fmt.Sprintf("%+v", value) == "" {
		return rtnErrs
	}

	options := strings.Split(validationData.Expected.(string), ";")
	var invalidValue interface{}

	switch obj.Kind() {
	case reflect.Array, reflect.Slice:
		var err error
		optionsVal := make(map[string]bool)
		for _, option := range options {
			option, err = v.loadExpectedValue(context, option)
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					return rtnErrs
				} else {
					continue
				}
			}
			optionsVal[option] = true
		}

		for i := 0; i < obj.Len(); i++ {
			nextValue := obj.Index(i)

			if !nextValue.CanInterface() {
				continue
			}

			_, ok := optionsVal[fmt.Sprintf("%+v", nextValue.Interface())]
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
		for _, option := range options {
			values := strings.Split(option, ":")
			if len(values) != 2 {
				continue
			}

			var err error
			values[1], err = v.loadExpectedValue(context, values[1])
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					return rtnErrs
				} else {
					continue
				}
			}

			optionsMap[values[0]] = values[1]
		}

		for _, key := range obj.MapKeys() {
			nextValue := obj.MapIndex(key)

			if !nextValue.CanInterface() {
				continue
			}

			val, ok := optionsMap[fmt.Sprintf("%+v", key.Interface())]
			if !ok || fmt.Sprintf("%+v", nextValue.Interface()) != fmt.Sprintf("%+v", val) {
				invalidValue = fmt.Sprintf("%+v:%+v", key.Interface(), nextValue.Interface())
				err := fmt.Errorf("the value [%+v] is different of the expected options [%+v] on field [%s]", nextValue.Interface(), validationData.Expected, validationData.Name)
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					break
				}
			}
		}

	default:
		var err error
		optionsVal := make(map[string]bool)
		for _, option := range options {
			option, err = v.loadExpectedValue(context, option)
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				if !v.validateAll {
					return rtnErrs
				} else {
					continue
				}
			}
			optionsVal[option] = true
		}

		_, ok := optionsVal[fmt.Sprintf("%+v", value)]
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

	isNull, obj, value := v.getValue(validationData.Value)
	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	size, e := strconv.Atoi(expected)
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
		if isNull {
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

	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	isNull, obj, value := v.getValue(validationData.Value)
	min, e := strconv.Atoi(expected)
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
		valueSize = obj.Int()
	case reflect.Float32, reflect.Float64:
		valueSize = int64(obj.Float())
	case reflect.String:
		valueSize = int64(len(strings.TrimSpace(obj.String())))
	case reflect.Bool:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatBool(obj.Bool()))))
	default:
		if isNull {
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

	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	isNull, obj, value := v.getValue(validationData.Value)
	max, e := strconv.Atoi(expected)
	if e != nil {
		err := fmt.Errorf("the size [%s] is invalid on field [%s] value [%+v]", validationData.Expected, validationData.Name, value)
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
		if isNull {
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

	isNull, _, value := v.getValue(validationData.Value)
	if !isNull {
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

	isNull, obj, value := v.getValue(validationData.Value)
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
		if isNull {
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

	isNull, _, value := v.getValue(validationData.Value)
	val := fmt.Sprintf("%+v", value)
	if isNull || val == "" {
		return rtnErrs
	}

	r, err := regexp.Compile(validationData.Expected.(string))
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if len(val) > 0 {
		if !r.MatchString(val) {
			err := fmt.Errorf("invalid value [%s] on field [%+v]", value, validationData.Name)
			rtnErrs = append(rtnErrs, err)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_special(context *ValidatorContext, validationData *ValidationData) []error {

	rtnErrs := make([]error, 0)

	isNull, _, value := v.getValue(validationData.Value)
	strValue := fmt.Sprintf("%+v", value)
	if isNull || strValue == "" {
		return rtnErrs
	}

	switch validationData.Expected {
	case ConstSpecialTagForDateDefault:
		validationData.Expected = ConstRegexForDateDefault
	case ConstSpecialTagForDateDDMMYYYY:
		validationData.Expected = ConstRegexForDateDDMMYYYY
	case ConstSpecialTagForDateYYYYMMDD:
		validationData.Expected = ConstRegexForDateYYYYMMDD
	case ConstSpecialTagForTimeDefault:
		validationData.Expected = ConstRegexForTimeDefault
	case ConstSpecialTagForTimeHHMMSS:
		validationData.Expected = ConstRegexForTimeHHMMSS
	case ConstSpecialTagForURL:
		validationData.Expected = ConstRegexForURL
	case ConstSpecialTagForEmail:
		validationData.Expected = ConstRegexForEmail
	default:
		err := fmt.Errorf("invalid special [%s] on field [%+v] value [%+v]", validationData.Expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	return v.validate_regex(context, validationData)
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
				err := context.execute(data.typ, data.obj, strings.Split(query, " "), &errs)

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

func (v *Validator) validate_string(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, _ := v.getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%+v]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	function := fmt.Sprintf("%+v", validationData.Expected)
	switch function {
	case ConstSetTagForTrim:
		v.string_trim(context, validationData)
	case ConstSetTagForTitle:
		v.string_title(context, validationData)
	case ConstSetTagForLower:
		v.string_lower(context, validationData)
	case ConstSetTagForUpper:
		v.string_upper(context, validationData)
	case ConstSetTagForKey:
		v.validate_key(context, validationData)
	default:
	}

	return rtnErrs
}

func (v *Validator) validate_set(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v.getValue(validationData.Value)
	if !obj.CanAddr() {
		err := fmt.Errorf("the object should be passed as a pointer! when validating field [%+v]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	newExpected := fmt.Sprintf("%+v", validationData.Expected)
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

				setValue(kind, value, newValue.obj.Interface())
			} else {
				err := fmt.Errorf("invalid set tag [%s] on field [%+v]", validationData.Expected, validationData.Name)
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

func (v *Validator) encode_md5(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)
	_, obj, value := v.getValue(validationData.Value)
	expected := fmt.Sprintf("%+v", value)
	kind := reflect.TypeOf(value).Kind()

	newValue := fmt.Sprintf("%x", md5.Sum([]byte(expected)))
	setValue(kind, obj, newValue)

	return rtnErrs
}

func (v *Validator) validate_key(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v.getValue(validationData.Value)
	if obj.CanAddr() {
		kind := reflect.TypeOf(value).Kind()

		switch kind {
		case reflect.String:
			expected, err := v.loadExpectedValue(context, validationData.Expected)
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				return rtnErrs
			}

			if expected == "" {
				expected = fmt.Sprintf("%+v", value)
			}

			setValue(kind, obj, convertToKey(strings.TrimSpace(expected), true))
		}
	}

	return rtnErrs
}

func (v *Validator) encode_random(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v.getValue(validationData.Value)

	if obj.CanAddr() {
		kind := reflect.TypeOf(value).Kind()

		switch obj.Kind() {
		case reflect.String:
			strValue := fmt.Sprintf("%+v", value)
			newValue := v.generate_random(strValue)

			setValue(kind, obj, newValue)
		default:
			return rtnErrs
		}
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

	_, parentObj, parentValue := v.getValue(validationData.Parent)

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

	isNil, _, value := v.getValue(validationData.Value)
	expected := fmt.Sprintf("%+v", value)

	if expected == "" || isNil {
		return rtnErrs
	}

	for _, r := range expected {
		if !unicode.IsLetter(r) {
			err := fmt.Errorf("the value [%+v] is invalid for type alphanumeric on field [%s] value [%+v]", expected, validationData.Name, value)
			rtnErrs = append(rtnErrs, err)
			break
		}
	}

	return rtnErrs
}

func (v *Validator) validate_numeric(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v.getValue(validationData.Value)
	expected := fmt.Sprintf("%+v", value)

	if expected == "" || isNil {
		return rtnErrs
	}

	for _, r := range expected {
		if !unicode.IsNumber(r) {
			err := fmt.Errorf("the value [%+v] is invalid for type numeric on field [%s] value [%+v]", expected, validationData.Name, value)
			rtnErrs = append(rtnErrs, err)
			break
		}
	}

	return rtnErrs
}

func (v *Validator) validate_bool(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, _, value := v.getValue(validationData.Value)
	expected := fmt.Sprintf("%+v", value)

	if expected == "" || isNil {
		return rtnErrs
	}

	switch strings.ToLower(expected) {
	case "true", "false":
	default:
		err := fmt.Errorf("the value [%+v] is invalid for type bool on field [%s] value [%+v]", expected, validationData.Name, value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) string_trim(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v.getValue(validationData.Value)

	if obj.CanAddr() {
		kind := reflect.TypeOf(value).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.TrimSpace(value.(string))
			regx := regexp.MustCompile("  +")
			newValue = string(regx.ReplaceAll(bytes.TrimSpace([]byte(newValue)), []byte(" ")))
			setValue(kind, obj, newValue)
		}
	}

	return rtnErrs
}

func (v *Validator) string_title(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v.getValue(validationData.Value)

	if obj.CanAddr() {
		kind := reflect.TypeOf(value).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.Title(value.(string))
			setValue(kind, obj, newValue)
		}
	}

	return rtnErrs
}

func (v *Validator) string_upper(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v.getValue(validationData.Value)

	if obj.CanAddr() {
		kind := reflect.TypeOf(value).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.ToUpper(value.(string))
			setValue(kind, obj, newValue)
		}
	}

	return rtnErrs
}

func (v *Validator) string_lower(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, value := v.getValue(validationData.Value)

	if obj.CanAddr() {
		kind := reflect.TypeOf(value).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.ToLower(value.(string))
			setValue(kind, obj, newValue)
		}
	}

	return rtnErrs
}

func (v *Validator) generate_random(strValue string) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz")

	newValue := []rune(strValue)

	for i, char := range newValue {
		if !unicode.IsSpace(char) {
			newValue[i] = chars[rand.Intn(len(chars))]
		}
	}

	return string(newValue)
}

func (v *Validator) validate_encode(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	_, obj, _ := v.getValue(validationData.Value)

	encoding := strings.ToLower(validationData.Expected.(string))

	if obj.CanAddr() {
		switch encoding {
		case ConstEncodeMd5:
			v.encode_md5(context, validationData)
		case ConstEncodeRandom:
			v.encode_random(context, validationData)
		default:
			err := fmt.Errorf("the encoding [%s] is invalid on field [%s]", encoding, validationData.Name)
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
