package validator

import (
	"reflect"
	"strings"
)

func (v *Validator) validate_options(context *ValidatorContext, validationData *ValidationData) []error {
	rtnErrs := make([]error, 0)

	isNil, obj, value := v._getValue(validationData.Value)
	strValue := v._convertToString(value)
	if isNil || strValue == "" {
		return rtnErrs
	}

	options := strings.Split(validationData.Expected.(string), constTagSplitValues)

	switch obj.Kind() {
	case reflect.Array, reflect.Slice:
		var err error
		var opt interface{}
		optionsVal := make(map[string]bool)
		for _, option := range options {
			opt, err = v._loadExpectedValue(context, option)
			if err != nil {
				rtnErrs = append(rtnErrs, err)
				if !v.canValidateAll {
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
				rtnErrs = append(rtnErrs, ErrorInvalidValue)
				if !v.canValidateAll {
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
				if !v.canValidateAll {
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
				rtnErrs = append(rtnErrs, ErrorInvalidValue)
				if !v.canValidateAll {
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
				if !v.canValidateAll {
					return rtnErrs
				} else {
					continue
				}
			}
			optionsVal[v._convertToString(opt)] = true
		}

		_, ok := optionsVal[v._convertToString(value)]
		if !ok {
			rtnErrs = append(rtnErrs, ErrorInvalidValue)
		}
	}

	return rtnErrs
}
