package validator

import (
	"fmt"
	"reflect"
	"strings"
)

func (v *Validator) newDefaultValues() defaultValues {
	return map[string]map[string]*data{
		ConstTagId:   make(map[string]*data),
		ConstTagJson: make(map[string]*data),
		ConstTagArg:  make(map[string]*data),
	}
}

func NewValidatorHandler(validator *Validator, args ...*Argument) *ValidatorContext {
	context := &ValidatorContext{
		validator: validator,
		values:    validator.newDefaultValues(),
	}

	for _, arg := range args {
		context.values[ConstTagArg][arg.Id] = &data{
			value: reflect.ValueOf(arg.Value),
			typ: reflect.StructField{
				Type: reflect.TypeOf(arg.Value),
			},
		}
	}

	return context
}

func (ctx *ValidatorContext) GetValue(tag string, id string) (*data, bool) {
	if values, ok := ctx.values[tag]; ok {
		if value, ok := values[id]; ok {
			return value, ok
		}
	}
	return nil, false
}

func (ctx *ValidatorContext) SetValue(tag string, id string, value *data) bool {
	if values, ok := ctx.values[tag]; ok {
		values[id] = value
		return true
	}
	return false
}

func (v *ValidatorContext) handleValidation(value interface{}) []error {
	var err error
	errs := make([]error, 0)

	// load id's
	if err = v.load(reflect.ValueOf(value), &errs); err != nil {
		return []error{err}
	}

	// execute
	if err = v.do(reflect.ValueOf(value), &errs); err != nil {
		return []error{err}
	}

	return errs
}

func (v *ValidatorContext) load(value reflect.Value, errs *[]error) error {
	types := reflect.TypeOf(value.Interface())

again:
	if (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) && !value.IsNil() {
		value = value.Elem()
		if value.IsValid() {
			types = value.Type()
			goto again
		} else {
			return nil
		}
	}

	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < types.NumField(); i++ {
			var dat *data
			nextValue := value.Field(i)
			nextType := types.Field(i)

			if !nextValue.CanInterface() {
				continue
			}

			tagValue, exists := nextType.Tag.Lookup(v.validator.tag)

			// save id sub tags
			if exists && strings.Contains(tagValue, fmt.Sprintf("%s=", ConstTagId)) {
				var id string

				split := strings.Split(tagValue, ",")
				var tag []string
				for _, item := range split {
					tag = strings.Split(item, "=")
					tag[0] = strings.TrimSpace(tag[0])

					switch tag[0] {
					case ConstTagId:
						id = tag[1]
						if dat == nil {
							dat = &data{
								value: nextValue,
								typ:   nextType,
							}
						}
					case ConstTagSet:
						newStruct := reflect.New(value.Type()).Elem()
						newField := newStruct.Field(i)

						if !strings.Contains(tagValue, fmt.Sprintf("%s=", ConstTagIf)) {
							setValue(nextValue.Kind(), newField, tag[1])
						} else {
							setValue(nextValue.Kind(), newField, value.Field(i).String())
						}

						dat = &data{
							value: newField,
							typ:   nextType,
						}
					}
				}
				v.SetValue(tag[0], id, dat)
			}

			// save json tags
			tagValue, exists = nextType.Tag.Lookup(ConstTagJson)
			if exists && tagValue != "-" {
				split := strings.Split(tagValue, ",")
				dat = &data{
					value: nextValue,
					typ:   nextType,
				}
				v.SetValue(ConstTagJson, split[0], dat)
			}

			if err := v.load(nextValue, errs); err != nil {
				return err
			}
		}

	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			nextValue := value.Index(i)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.load(nextValue, errs); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range value.MapKeys() {
			nextValue := value.MapIndex(key)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.load(key, errs); err != nil {
				return err
			}
			if err := v.load(nextValue, errs); err != nil {
				return err
			}
		}

	default:
		// do nothing ...
	}
	return nil
}

func (v *ValidatorContext) do(value reflect.Value, errs *[]error) error {
	types := reflect.TypeOf(value.Interface())

again:
	if (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) && !value.IsNil() {
		value = value.Elem()
		if value.IsValid() {
			types = value.Type()
			goto again
		} else {
			return nil
		}
	}

	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < types.NumField(); i++ {
			nextValue := value.Field(i)
			nextType := types.Field(i)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.doValidate(nextValue, nextType, errs); err != nil {
				return err
			}

			if len(*errs) > 0 && !v.validator.validateAll {
				return nil
			}

			if err := v.do(nextValue, errs); err != nil {
				return err
			}

			if len(*errs) > 0 && !v.validator.validateAll {
				return nil
			}
		}

	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			nextValue := value.Index(i)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.do(nextValue, errs); err != nil {
				return err
			}

			if len(*errs) > 0 && !v.validator.validateAll {
				return nil
			}
		}

	case reflect.Map:
		for _, key := range value.MapKeys() {
			nextValue := value.MapIndex(key)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.do(key, errs); err != nil {
				return err
			}

			if len(*errs) > 0 && !v.validator.validateAll {
				return nil
			}

			if err := v.do(nextValue, errs); err != nil {
				return err
			}

			if len(*errs) > 0 && !v.validator.validateAll {
				return nil
			}
		}

	default:
		// do nothing ...
	}
	return nil
}

func (v *ValidatorContext) doValidate(value reflect.Value, typ reflect.StructField, errs *[]error) error {

	tag, exists := typ.Tag.Lookup(v.validator.tag)
	if !exists {
		return nil
	}

	validations := strings.Split(tag, ",")

	return v.execute(typ, value, validations, errs)
}

func (v *ValidatorContext) getFieldId(validations []string) string {
	for _, validation := range validations {
		options := strings.SplitN(validation, "=", 2)
		tag := strings.TrimSpace(options[0])

		if tag == ConstTagId {
			return options[1]
		}
	}

	return ""
}

func (v *ValidatorContext) execute(typ reflect.StructField, value reflect.Value, validations []string, errs *[]error) error {
	var err error
	var itErrs []error
	var replacedErrors = make(map[error]bool)
	skipValidation := false
	onlyHandleNextErrorTag := false

	defer func(){
		*errs = append(*errs, itErrs...)
	}()

	baseData := &BaseData{
		Id:        v.getFieldId(validations),
		Arguments: make([]interface{}, 0),
	}

	for _, validation := range validations {
		var name string
		var tag string
		var prefix string

		options := strings.SplitN(validation, "=", 2)
		tag = strings.TrimSpace(options[0])

		if split := strings.SplitN(tag, ":", 2); len(split) > 1 {
			prefix = split[0]
			tag = split[1]
		}

		if onlyHandleNextErrorTag && !v.validator.validateAll && tag != ConstTagError {
			continue
		}

		if _, ok := v.validator.activeHandlers[tag]; !ok {
			return fmt.Errorf("invalid tag [%s]", tag)
		}

		var expected interface{}
		if len(options) > 1 {
			expected = strings.TrimSpace(options[1])
		}

		jsonName, exists := typ.Tag.Lookup("json")
		if exists {
			split := strings.SplitN(jsonName, ",", 2)
			name = split[0]
		} else {
			name = typ.Name
		}

		if skipValidation {
			if tag == ConstTagIf {
				skipValidation = false
			} else {
				continue
			}
		}

		// execute validations
		switch prefix {
		case ConstPrefixTagKey, ConstPrefixTagItem:
			types := reflect.TypeOf(value.Interface())

			if !value.CanInterface() {
				return nil
			}

		again:
			if (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) && !value.IsNil() {
				value = value.Elem()
				if value.IsValid() {
					types = value.Type()
					goto again
				} else {
					return nil
				}
			}

			if prefix == ConstPrefixTagKey && value.Kind() != reflect.Map {
				continue
			}

			switch value.Kind() {
			case reflect.Array, reflect.Slice:
				for i := 0; i < value.Len(); i++ {
					nextValue := value.Index(i)

					if !nextValue.CanInterface() {
						continue
					}

					validationData := ValidationData{
						BaseData:       baseData,
						Name:           name,
						Field:          typ.Name,
						Parent:         value,
						Value:          nextValue,
						Expected:       expected,
						Errors:         &itErrs,
						ErrorsReplaced: replacedErrors,
					}

					err = v.executeHandlers(tag, &validationData, &itErrs)
				}
			case reflect.Map:
				for _, key := range value.MapKeys() {

					var nextValue reflect.Value

					switch prefix {
					case ConstPrefixTagKey:
						nextValue = key
					case ConstPrefixTagItem:
						nextValue = value.MapIndex(key)
					}

					if !key.CanInterface() {
						continue
					}

					validationData := ValidationData{
						BaseData:       baseData,
						Name:           name,
						Field:          typ.Name,
						Parent:         value,
						Value:          nextValue,
						Expected:       expected,
						Errors:         &itErrs,
						ErrorsReplaced: replacedErrors,
					}

					err = v.executeHandlers(tag, &validationData, &itErrs)
				}
			case reflect.Struct:
				for i := 0; i < types.NumField(); i++ {
					nextValue := value.Field(i)

					if !nextValue.CanInterface() {
						continue
					}

					validationData := ValidationData{
						BaseData:       baseData,
						Name:           name,
						Field:          typ.Name,
						Parent:         value,
						Value:          nextValue,
						Expected:       expected,
						Errors:         &itErrs,
						ErrorsReplaced: replacedErrors,
					}

					err = v.executeHandlers(tag, &validationData, &itErrs)
				}
			}

		default:
			if prefix != "" {
				return fmt.Errorf("invalid tag prefix [%s] on tag [%s]", prefix, tag)
			}

			validationData := ValidationData{
				BaseData:       baseData,
				Name:           name,
				Field:          typ.Name,
				Parent:         value,
				Value:          value,
				Expected:       expected,
				Errors:         &itErrs,
				ErrorsReplaced: replacedErrors,
			}

			err = v.executeHandlers(tag, &validationData, &itErrs)
		}

		if onlyHandleNextErrorTag && !v.validator.validateAll && tag == ConstTagError {
			if err == ErrorSkipValidation {
				skipValidation = true
				continue
			}

			return nil
		}

		if err != nil {
			if err == ErrorSkipValidation {
				skipValidation = true
				continue
			} else {
				return err
			}
		}

		if len(*errs) > 0 {
			if !onlyHandleNextErrorTag && !v.validator.validateAll && tag != ConstTagError {
				onlyHandleNextErrorTag = true
				continue
			}

			if !v.validator.validateAll {
				return nil
			}
		}
	}

	return nil
}

func (v *ValidatorContext) executeHandlers(tag string, validationData *ValidationData, errs *[]error) error {
	var err error

	if _, ok := v.validator.handlersBefore[tag]; ok {
		if rtnErrs := v.validator.handlersBefore[tag](v, validationData); rtnErrs != nil && len(rtnErrs) > 0 {

			// skip validation
			if rtnErrs[0] == ErrorSkipValidation {
				return rtnErrs[0]
			}
			*errs = append(*errs, rtnErrs...)
		}
	}

	if _, ok := v.validator.handlersMiddle[tag]; ok {
		if rtnErrs := v.validator.handlersMiddle[tag](v, validationData); rtnErrs != nil && len(rtnErrs) > 0 {
			*errs = append(*errs, rtnErrs...)
		}
	}

	if _, ok := v.validator.handlersAfter[tag]; ok {
		if rtnErrs := v.validator.handlersAfter[tag](v, validationData); rtnErrs != nil && len(rtnErrs) > 0 {
			*errs = append(*errs, rtnErrs...)
		}
	}

	return err
}
