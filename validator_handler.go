package validator

import (
	"fmt"
	"reflect"
	"strings"
)

func NewValidatorHandler(validator *Validator) *ValidatorContext {
	return &ValidatorContext{
		validator: validator,
		Values:    make(map[string]*Data),
	}
}
func (v *ValidatorContext) handleValidation(value interface{}) []error {
	errs := make([]error, 0)

	// load id's
	v.load(reflect.ValueOf(value), &errs)

	// execute
	v.do(reflect.ValueOf(value), &errs)

	return errs
}

func (v *ValidatorContext) load(value reflect.Value, errs *[]error) error {
	types := reflect.TypeOf(value.Interface())

	if !value.CanInterface() {
		return nil
	}

	if value.Kind() == reflect.Ptr && !value.IsNil() {
		value = value.Elem()

		if value.IsValid() {
			types = value.Type()
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

			tagValue, exists := nextType.Tag.Lookup(v.validator.tag)
			if !exists || strings.Contains(tagValue, fmt.Sprintf("%s=", ConstTagId)) {
				var id string
				var data *Data

				split := strings.Split(tagValue, ",")
				for _, item := range split {
					tag := strings.Split(item, "=")

					switch strings.TrimSpace(tag[0]) {
					case ConstTagId:
						id = tag[1]
						if data == nil {
							data = &Data{
								Obj:   nextValue,
								Type:  nextType,
								IsSet: false,
							}
						}
					case ConstTagSet:
						isSet := false
						newStruct := reflect.New(value.Type()).Elem()
						newField := newStruct.Field(i)

						if !strings.Contains(tagValue, fmt.Sprintf("%s=", ConstTagIf)) {
							isSet = true
							setValue(nextValue.Kind(), newField, tag[1])
						} else {
							setValue(nextValue.Kind(), newField, value.Field(i).String())
						}

						data = &Data{
							Obj:   newField,
							Type:  nextType,
							IsSet: isSet,
						}
					}
				}
				v.Values[id] = data
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

	if !value.CanInterface() {
		return nil
	}

	if value.Kind() == reflect.Ptr && !value.IsNil() {
		value = value.Elem()

		if value.IsValid() {
			types = value.Type()
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

				if !v.validator.validateAll {
					return err
				}
			}

			if err := v.do(nextValue, errs); err != nil {
				if !v.validator.validateAll {
					return err
				}
			}
		}

	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			nextValue := value.Index(i)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.do(nextValue, errs); err != nil {
				if !v.validator.validateAll {
					return err
				}
			}
		}

	case reflect.Map:
		for _, key := range value.MapKeys() {
			nextValue := value.MapIndex(key)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.do(key, errs); err != nil {
				if !v.validator.validateAll {
					return err
				}
			}
			if err := v.do(nextValue, errs); err != nil {
				if !v.validator.validateAll {
					return err
				}
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

		if _, ok := v.validator.activeHandlers[tag]; !ok {
			err := fmt.Errorf("invalid tag [%s]", tag)
			*errs = append(*errs, err)

			if !v.validator.validateAll {
				return err
			}
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

		// execute validations
		switch prefix {
		case ConstPrefixTagKey, ConstPrefixTagItem:
			types := reflect.TypeOf(value.Interface())

			if !value.CanInterface() {
				return nil
			}

			if value.Kind() == reflect.Ptr && !value.IsNil() {
				value = value.Elem()

				if value.IsValid() {
					types = value.Type()
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

					if !v.validator.validateAll {
						return err
					}
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

					if !v.validator.validateAll {
						return err
					}
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

					if !v.validator.validateAll {
						return err
					}
				}
			}

		default:
			if prefix != "" {
				err := fmt.Errorf("invalid tag prefix [%s] on tag [%s]", prefix, tag)
				itErrs = append(itErrs, err)

				if !v.validator.validateAll {
					return err
				}
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

		if !v.validator.validateAll {
			return err
		}
	}

	*errs = append(*errs, itErrs...)

	return err
}

func (v *ValidatorContext) executeHandlers(tag string, validationData *ValidationData, errs *[]error) error {
	var err error

	if _, ok := v.validator.handlersBefore[tag]; ok {
		if rtnErrs := v.validator.handlersBefore[tag](v, validationData); rtnErrs != nil && len(rtnErrs) > 0 {

			// skip validation
			if rtnErrs[0] == ErrorSkipValidation {
				return nil
			}
			*errs = append(*errs, rtnErrs...)
			err = rtnErrs[0]
		}
	}

	if _, ok := v.validator.handlersMiddle[tag]; ok {
		if rtnErrs := v.validator.handlersMiddle[tag](v, validationData); rtnErrs != nil && len(rtnErrs) > 0 {
			*errs = append(*errs, rtnErrs...)
			err = rtnErrs[0]
		}
	}

	if _, ok := v.validator.handlersAfter[tag]; ok {
		if rtnErrs := v.validator.handlersAfter[tag](v, validationData); rtnErrs != nil && len(rtnErrs) > 0 {
			*errs = append(*errs, rtnErrs...)
			err = rtnErrs[0]
		}
	}

	return err
}
