package validator

import (
	"reflect"

	"github.com/joaosoft/logger"
)

func (v *Validator) init() {
	v.handlersBefore = v.NewDefaultBeforeHandlers()
	v.handlersMiddle = v.NewDefaultMiddleHandlers()
	v.handlersAfter = v.NewDefaultPosHandlers()
	v.activeHandlers = v.NewActiveHandlers()

}

type Validator struct {
	tag              string
	activeHandlers   map[string]bool
	handlersBefore   map[string]BeforeTagHandler
	handlersMiddle   map[string]MiddleTagHandler
	handlersAfter    map[string]AfterTagHandler
	errorCodeHandler ErrorCodeHandler
	callbacks        map[string]CallbackHandler
	sanitize         []string
	logger           logger.ILogger
	validateAll      bool
}

type ErrorCodeHandler func(context *ValidatorContext, validationData *ValidationData, args ...interface{}) error
type CallbackHandler func(context *ValidatorContext, validationData *ValidationData, args ...interface{}) []error

type BeforeTagHandler func(context *ValidatorContext, validationData *ValidationData, args ...interface{}) []error
type MiddleTagHandler func(context *ValidatorContext, validationData *ValidationData, args ...interface{}) []error
type AfterTagHandler func(context *ValidatorContext, validationData *ValidationData, args ...interface{}) []error

type ValidatorContext struct {
	validator *Validator
	Values    map[string]*Data
}

type BaseData struct {
	Id        string
	Arguments []interface{}
}

type ValidationData struct {
	*BaseData
	Code           string
	Field          string
	Parent         reflect.Value
	Value          reflect.Value
	Name           string
	Expected       interface{}
	ErrorData      *ErrorData
	Errors         *[]error
	ErrorsReplaced map[error]bool
}

type ErrorData struct {
	Code      string
	Arguments []interface{}
}

type Data struct {
	Obj   reflect.Value
	Type  reflect.StructField
	IsSet bool
}

type Expression struct {
	Data         *Data
	Result       error
	Expected     string
	NextOperator Operator
}
