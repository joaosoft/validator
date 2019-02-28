package validator

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"encoding/json"
	"errors"

	"github.com/satori/go.uuid"
)

func (v *Validator) loadExpectedValue(context *ValidatorContext, expected interface{}) (string, error) {
	newExpected := fmt.Sprintf("%+v", expected)
	if matched, err := regexp.MatchString(ConstRegexForTagValue, newExpected); err != nil {
		return "", err
	} else {
		if matched {
			replacer := strings.NewReplacer("{", "", "}", "")
			id := replacer.Replace(newExpected)
			newExpected = fmt.Sprintf("%+v", context.Values[id].Value.Interface())
		}
	}

	return newExpected, nil
}

func (v *Validator) validate_value(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if fmt.Sprintf("%+v", validationData.Value) == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if fmt.Sprintf("%+v", validationData.Value) != expected {
		err := fmt.Errorf("the value [%+v] is different of the expected [%+v] on field [%s] value [%+v]", validationData.Value, expected, validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_sanitize(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if fmt.Sprintf("%+v", validationData.Value) == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	val := fmt.Sprintf("%+v", validationData.Value)
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
		err := fmt.Errorf("the value [%+v] is has invalid characters [%+v] on field [%s]", validationData.Value, strings.Join(invalid, ","), validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_not(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if fmt.Sprintf("%+v", validationData.Value) == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if fmt.Sprintf("%+v", validationData.Value) == fmt.Sprintf("%+v", expected) {
		err := fmt.Errorf("the expected [%+v] should be different of the [%+v] on field [%s]", expected, validationData.Value, validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_options(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if fmt.Sprintf("%+v", validationData.Value) == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	options := strings.Split(validationData.Expected.(string), ";")
	var invalidValue interface{}

	switch validationData.Value.Kind() {
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

		for i := 0; i < validationData.Value.Len(); i++ {
			nextValue := validationData.Value.Index(i)

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

		for _, key := range validationData.Value.MapKeys() {
			nextValue := validationData.Value.MapIndex(key)

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

		_, ok := optionsVal[fmt.Sprintf("%+v", validationData.Value)]
		if !ok {
			invalidValue = validationData.Value
			err := fmt.Errorf("the value [%+v] is different of the expected options [%+v] on field [%s]", invalidValue, validationData.Expected, validationData.Name)
			rtnErrs = append(rtnErrs, err)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_size(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	expected, err := v.loadExpectedValue(context, validationData.Expected)
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	size, e := strconv.Atoi(expected)
	if e != nil {
		err := fmt.Errorf("the size [%s] is invalid on field [%s] value [%+v]", expected, validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	var valueSize int64

	switch validationData.Value.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		valueSize = int64(validationData.Value.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueSize = int64(len(strings.TrimSpace(strconv.Itoa(int(validationData.Value.Int())))))
	case reflect.Float32, reflect.Float64:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatFloat(validationData.Value.Float(), 'g', 1, 64))))
	case reflect.String:
		valueSize = int64(len(strings.TrimSpace(validationData.Value.String())))
	case reflect.Bool:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatBool(validationData.Value.Bool()))))
	default:
		if validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil() {
			break
		}
		valueSize = int64(len(strings.TrimSpace(validationData.Value.String())))
	}

	if valueSize != int64(size) {
		err := fmt.Errorf("the length [%+v] is lower then the expected [%+v] on field [%s] value [%+v]", valueSize, expected, validationData.Name, validationData.Value)
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

	min, e := strconv.Atoi(expected)
	if e != nil {
		err := fmt.Errorf("the size [%s] is invalid on field [%s] value [%+v]", expected, validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	var valueSize int64

	switch validationData.Value.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		valueSize = int64(validationData.Value.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueSize = validationData.Value.Int()
	case reflect.Float32, reflect.Float64:
		valueSize = int64(validationData.Value.Float())
	case reflect.String:
		valueSize = int64(len(strings.TrimSpace(validationData.Value.String())))
	case reflect.Bool:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatBool(validationData.Value.Bool()))))
	default:
		if validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil() {
			break
		}
		valueSize = int64(len(strings.TrimSpace(validationData.Value.String())))
	}

	if valueSize < int64(min) {
		err := fmt.Errorf("the length [%+v] is lower then the expected [%+v] on field [%s] value [%+v]", valueSize, expected, validationData.Name, validationData.Value)
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

	max, e := strconv.Atoi(expected)
	if e != nil {
		err := fmt.Errorf("the size [%s] is invalid on field [%s] value [%+v]", validationData.Expected, validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	var valueSize int64

	switch validationData.Value.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		valueSize = int64(validationData.Value.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueSize = validationData.Value.Int()
	case reflect.Float32, reflect.Float64:
		valueSize = int64(validationData.Value.Float())
	case reflect.String:
		valueSize = int64(len(strings.TrimSpace(validationData.Value.String())))
	case reflect.Bool:
		valueSize = int64(len(strings.TrimSpace(strconv.FormatBool(validationData.Value.Bool()))))
	default:
		if validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil() {
			break
		}
		valueSize = int64(len(strings.TrimSpace(validationData.Value.String())))
	}

	if valueSize > int64(max) {
		err := fmt.Errorf("the length [%+v] is bigger then the expected [%+v] on field [%s] value [%+v]", valueSize, expected, validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_nonzero(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if errs := v.validate_iszero(context, validationData); len(errs) == 0 {
		err := fmt.Errorf("the value shouldn't be zero on field [%s]", validationData.Name)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_isnull(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if validationData.Value.CanAddr() {
		err := fmt.Errorf("the value should be null on field [%s] instead of [%+v]", validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_nonnull(context *ValidatorContext, validationData *ValidationData) []error {
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

	switch validationData.Value.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:

		switch validationData.Value.Type() {
		case reflect.TypeOf(uuid.UUID{}):
			if validationData.Value.Interface().(uuid.UUID) == uuid.Nil {
				isZero = true
			}
		default:
			isZero = validationData.Value.Len() == 0
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		isZero = validationData.Value.Int() == 0
	case reflect.Float32, reflect.Float64:
		isZero = validationData.Value.Float() == 0
	case reflect.String:
		isZero = len(strings.TrimSpace(validationData.Value.String())) == 0
	case reflect.Bool:
		isZero = validationData.Value.Bool() == false
	case reflect.Struct:
		if validationData.Value.Interface() == reflect.Zero(validationData.Value.Type()).Interface() {
			isZero = true
		}
	default:
		if validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil() {
			isZero = true
		}
	}

	if !isZero {
		err := fmt.Errorf("the value should be zero on field [%s] instead of [%+v]", validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) validate_regex(context *ValidatorContext, validationData *ValidationData) []error {

	rtnErrs := make([]error, 0)

	if validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil() {
		return rtnErrs
	}

	if fmt.Sprintf("%+v", validationData.Value) == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	r, err := regexp.Compile(validationData.Expected.(string))
	if err != nil {
		rtnErrs = append(rtnErrs, err)
		return rtnErrs
	}

	if len(fmt.Sprintf("%+v", validationData.Value)) > 0 {
		if !r.MatchString(fmt.Sprintf("%+v", validationData.Value)) {
			err := fmt.Errorf("invalid data [%s] on field [%+v] value [%+v]", validationData.Value, validationData.Name, validationData.Value)
			rtnErrs = append(rtnErrs, err)
		}
	}

	return rtnErrs
}

func (v *Validator) validate_special(context *ValidatorContext, validationData *ValidationData) []error {

	rtnErrs := make([]error, 0)

	if fmt.Sprintf("%+v", validationData.Value) == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
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
		err := fmt.Errorf("invalid special [%s] on field [%+v] value [%+v]", validationData.Expected, validationData.Name, validationData.Value)
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
			if matched, err := regexp.MatchString(ConstRegexForTagValue, validationData.Expected.(string)); err != nil {
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
					messageBytes, _ := json.Marshal(Error{
						Code:    fmt.Sprintf("%+v", validationData.Expected),
						Message: (*validationData.Errors)[i].Error(),
					})
					newErr := errors.New(string(messageBytes))
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
	var expressions []*Expression
	var expression *Expression
	var query string

	// read conditions
	size := len(str)

	for i := 0; i < size; i++ {
		switch str[i] {
		case '(':
			continue

		case ')':
			start := strings.Index(query, "id=")
			if start == -1 {
				return rtnErrs
			}

			end := strings.Index(query[start:], " ")
			if end == -1 {
				end = size - 1
			}

			id := query[start+3 : end]
			query = query[end+1:]

			if data, ok := context.Values[id]; ok {
				var errs []error
				err := context.execute(data.Value, data.Type, data.Obj, data.MutableObj, strings.Split(query, " "), &errs)

				// get next operator
				var operator Operator
				if index := strings.Index(str[i+1:], "("); index > -1 {
					operator = Operator(strings.TrimSpace(str[i+1 : i+1+index]))

					str = str[i+1+index:]
					i = 0
					size = len(str)
				}

				expression = &Expression{
					Data:         data,
					Result:       err,
					NextOperator: operator,
					Expected:     query,
				}
				expressions = append(expressions, expression)
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
			if expr.Result == nil {
				condition = "ok"
			} else {
				condition = "ko"
			}
		} else {

			switch prevOperator {
			case AND:
				if expr.Result != nil {
					condition = "ko"
				}
			case OR:
				if expr.Result == nil && condition == "ko" {
					condition = "ok"
				}
			case NONE:
				if expr.Result == nil {
					condition = "ok"
				}
			}
		}

		prevOperator = expr.NextOperator
	}

	if condition == "ko" {
		return []error{ErrorSkipValidation}
	}

	return nil
}

func (v *Validator) validate_set(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if !validationData.MutableObj.CanAddr() {
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
			validationData.Expected = validationData.Value.Interface()

			switch id {
			case ConstSetTagForTrim:
				v.set_trim(context, validationData)
			case ConstSetTagForTitle:
				v.set_title(context, validationData)
			case ConstSetTagForLower:
				v.set_lower(context, validationData)
			case ConstSetTagForUpper:
				v.set_upper(context, validationData)
			case ConstSetTagForKey:
				v.set_key(context, validationData)
			default:
				if newValue, ok := context.Values[id]; ok {
					value := validationData.MutableObj.FieldByName(validationData.Field)
					kind := reflect.TypeOf(value.Interface()).Kind()

					setValue(kind, value, newValue.Value.Interface())
				} else {
					err := fmt.Errorf("invalid set tag [%s] on field [%+v]", validationData.Expected, validationData.Name)
					rtnErrs = append(rtnErrs, err)
					return rtnErrs
				}
			}
		} else {
			value := validationData.MutableObj.FieldByName(validationData.Field)
			kind := reflect.TypeOf(value.Interface()).Kind()

			setValue(kind, value, validationData.Expected)
		}
	}

	return rtnErrs
}

func (v *Validator) set_key(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if validationData.MutableObj.CanAddr() {
		value := validationData.MutableObj.FieldByName(validationData.Field)
		kind := reflect.TypeOf(value.Interface()).Kind()

		switch kind {
		case reflect.String:
			expected, err := v.loadExpectedValue(context, validationData.Expected)
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				return rtnErrs
			}

			setValue(kind, value, convertToKey(strings.TrimSpace(expected), true))
		}
	}

	return rtnErrs
}

func setValue(kind reflect.Kind, mutable reflect.Value, newValue interface{}) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, _ := strconv.Atoi(newValue.(string))
		mutable.SetInt(int64(v))
	case reflect.Float32, reflect.Float64:
		v, _ := strconv.ParseFloat(newValue.(string), 64)
		mutable.SetFloat(v)
	case reflect.String:
		mutable.SetString(newValue.(string))
	case reflect.Bool:
		v, _ := strconv.ParseBool(newValue.(string))
		mutable.SetBool(v)
	}
}

func (v *Validator) validate_distinct(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if validationData.MutableObj.CanAddr() {
		value := validationData.MutableObj.FieldByName(validationData.Field)
		kind := reflect.TypeOf(value.Interface()).Kind()

		if kind != reflect.Array && kind != reflect.Slice {
			return rtnErrs
		}
		newInstance := reflect.New(value.Type()).Elem()

		values := make(map[interface{}]bool)
		for i := 0; i < value.Len(); i++ {

			indexValue := value.Index(i)
			if indexValue.Kind() == reflect.Ptr && !indexValue.IsNil() {
				indexValue = value.Index(i).Elem()
			}

			if _, ok := values[indexValue.Interface()]; ok {
				continue
			}

			switch indexValue.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Float32, reflect.Float64,
				reflect.String,
				reflect.Bool:
				if value.Index(i).Kind() == reflect.Ptr && !value.Index(i).IsNil() {
					newInstance = reflect.Append(newInstance, indexValue.Addr())
				} else {
					newInstance = reflect.Append(newInstance, indexValue)
				}

				values[indexValue.Interface()] = true
			}
		}

		// set the new instance without duplicated values
		value.Set(newInstance)
	}

	return rtnErrs
}

func (v *Validator) validate_alpha(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	expected := fmt.Sprintf("%+v", validationData.Value)

	if expected == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	for _, r := range expected {
		if !unicode.IsLetter(r) {
			err := fmt.Errorf("the value [%+v] is invalid for type alphanumeric on field [%s] value [%+v]", expected, validationData.Name, validationData.Value)
			rtnErrs = append(rtnErrs, err)
			break
		}
	}

	return rtnErrs
}

func (v *Validator) validate_numeric(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)
	expected := fmt.Sprintf("%+v", validationData.Value)

	if expected == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	for _, r := range expected {
		if !unicode.IsNumber(r) {
			err := fmt.Errorf("the value [%+v] is invalid for type numeric on field [%s] value [%+v]", expected, validationData.Name, validationData.Value)
			rtnErrs = append(rtnErrs, err)
			break
		}
	}

	return rtnErrs
}

func (v *Validator) validate_bool(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)
	expected := fmt.Sprintf("%+v", validationData.Value)

	if expected == "" || (validationData.Value.Kind() == reflect.Ptr && validationData.Value.IsNil()) {
		return rtnErrs
	}

	switch strings.ToLower(expected) {
	case "true", "false":
	default:
		err := fmt.Errorf("the value [%+v] is invalid for type bool on field [%s] value [%+v]", expected, validationData.Name, validationData.Value)
		rtnErrs = append(rtnErrs, err)
	}

	return rtnErrs
}

func (v *Validator) set_trim(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if validationData.MutableObj.CanAddr() {
		value := validationData.MutableObj.FieldByName(validationData.Field)
		kind := reflect.TypeOf(value.Interface()).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.TrimSpace(value.Interface().(string))
			regx := regexp.MustCompile("  +")
			newValue = string(regx.ReplaceAll(bytes.TrimSpace([]byte(newValue)), []byte(" ")))
			setValue(kind, value, newValue)
		}
	}

	return rtnErrs
}

func (v *Validator) set_title(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if validationData.MutableObj.CanAddr() {
		value := validationData.MutableObj.FieldByName(validationData.Field)
		kind := reflect.TypeOf(value.Interface()).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.Title(value.Interface().(string))
			setValue(kind, value, newValue)
		}
	}

	return rtnErrs
}

func (v *Validator) set_upper(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if validationData.MutableObj.CanAddr() {
		value := validationData.MutableObj.FieldByName(validationData.Field)
		kind := reflect.TypeOf(value.Interface()).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.ToUpper(value.Interface().(string))
			setValue(kind, value, newValue)
		}
	}

	return rtnErrs
}

func (v *Validator) set_lower(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	if validationData.MutableObj.CanAddr() {
		value := validationData.MutableObj.FieldByName(validationData.Field)
		kind := reflect.TypeOf(value.Interface()).Kind()

		switch kind {
		case reflect.String:
			newValue := strings.ToLower(value.Interface().(string))
			setValue(kind, value, newValue)
		}
	}

	return rtnErrs
}
