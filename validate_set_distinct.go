package validator

import "reflect"

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
