package validator

import (
	"fmt"
	"reflect"
	"strings"
)

func NewValidatorHandler(validator *Validator) *ValidatorContext {
	return &ValidatorContext{
		validator: validator,
		values:    make(map[string]*Data),
	}
}
func (v *ValidatorContext) handleValidation(obj interface{}) []error {
	errs := make([]error, 0)
	mutable := reflect.ValueOf(obj)

	if mutable.Kind() == reflect.Ptr && !mutable.IsNil() {
		mutable = mutable.Elem()
	}
	// load id's

	v.load(obj, mutable, &errs)

	// execute
	v.do(obj, mutable, &errs)

	return errs
}

func (v *ValidatorContext) load(obj interface{}, mutable reflect.Value, errs *[]error) error {
	types := reflect.TypeOf(obj)
	value := reflect.ValueOf(obj)

	if !value.CanInterface() {
		return nil
	}

	fmt.Println(value.Kind())
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

			if nextValue.Kind() == reflect.Ptr && !nextValue.IsNil() {
				nextValue = nextValue.Elem()
			}

			if !nextValue.CanInterface() {
				continue
			}

			tagValue, exists := nextType.Tag.Lookup(v.validator.tag)
			if !exists || strings.Contains(tagValue, "id=") {
				var id string
				var data *Data

				split := strings.Split(tagValue, ",")
				for _, item := range split {
					tag := strings.Split(item, "=")

					switch strings.TrimSpace(tag[0]) {
					case "id":
						id = tag[1]
						if data == nil {
							data = &Data{
								Value:      nextValue,
								Obj:        &obj,
								MutableObj: nextValue,
								Type:       nextType,
							}
						}
					case "set":
						newStruct := reflect.New(value.Type()).Elem()
						newField := newStruct.Field(i)
						setValue(nextValue.Kind(), newField, tag[1])

						data = &Data{
							Value:      newField,
							Obj:        &obj,
							MutableObj: nextValue,
							Type:       nextType,
						}
					}
				}

				v.values[id] = data
			}

			if err := v.load(nextValue.Interface(), nextValue, errs); err != nil {
				return err
			}
		}

	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			nextValue := value.Index(i)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.load(nextValue.Interface(), nextValue, errs); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range value.MapKeys() {
			nextValue := value.MapIndex(key)

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.load(key.Interface(), key, errs); err != nil {
				return err
			}
			if err := v.load(nextValue.Interface(), nextValue, errs); err != nil {
				return err
			}
		}

	default:
		// do nothing ...
	}
	return nil
}

func (v *ValidatorContext) do(obj interface{}, mutable reflect.Value, errs *[]error) error {
	types := reflect.TypeOf(obj)
	value := reflect.ValueOf(obj)

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

			if nextValue.Kind() == reflect.Ptr && !nextValue.IsNil() {
				nextValue = nextValue.Elem()
			}

			if !nextValue.CanInterface() {
				continue
			}

			if err := v.doValidate(nextValue, nextType, obj, mutable, errs); err != nil {

				if !v.validator.validateAll {
					return err
				}
			}

			if err := v.do(nextValue.Interface(), nextValue, errs); err != nil {
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

			if err := v.do(nextValue.Interface(), nextValue, errs); err != nil {
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

			if err := v.do(key.Interface(), nextValue, errs); err != nil {
				if !v.validator.validateAll {
					return err
				}
			}
			if err := v.do(nextValue.Interface(), nextValue, errs); err != nil {
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

func (v *ValidatorContext) doValidate(value reflect.Value, typ reflect.StructField, obj interface{}, mutable reflect.Value, errs *[]error) error {

	tag, exists := typ.Tag.Lookup(v.validator.tag)
	if !exists {
		return nil
	}

	validations := strings.Split(tag, ",")

	return v.executeHandlers(value, typ, obj, mutable, validations, errs)
}

func (v *ValidatorContext) executeHandlers(value reflect.Value, typ reflect.StructField, obj interface{}, mutable reflect.Value, validations []string, errs *[]error) error {
	var err error
	var itErrs []error
	var replacedErrors = make(map[error]bool)

	for _, validation := range validations {
		var name string

		options := strings.SplitN(validation, "=", 2)
		tag := strings.TrimSpace(options[0])

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

		validationData := ValidationData{
			Name:           name,
			Value:          value,
			Field:          typ.Name,
			Obj:            obj,
			MutableObj:     mutable,
			Expected:       expected,
			Errors:         &itErrs,
			ErrorsReplaced: replacedErrors,
		}
		// execute validations
		if _, ok := v.validator.handlersBefore[tag]; ok {
			if rtnErrs := v.validator.handlersBefore[tag](v, &validationData); rtnErrs != nil && len(rtnErrs) > 0 {

				// skip validation
				if rtnErrs[0] == ErrorSkipValidation {
					return nil
				}
				itErrs = append(itErrs, rtnErrs...)
				err = rtnErrs[0]
			}
		}

		if _, ok := v.validator.handlersMiddle[tag]; ok {
			if rtnErrs := v.validator.handlersMiddle[tag](v, &validationData); rtnErrs != nil && len(rtnErrs) > 0 {
				itErrs = append(itErrs, rtnErrs...)
				err = rtnErrs[0]
			}
		}

		if _, ok := v.validator.handlersAfter[tag]; ok {
			if rtnErrs := v.validator.handlersAfter[tag](v, &validationData); rtnErrs != nil && len(rtnErrs) > 0 {
				itErrs = append(itErrs, rtnErrs...)
				err = rtnErrs[0]
			}
		}
	}

	*errs = append(*errs, itErrs...)

	return err
}
