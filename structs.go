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

type ErrorCodeHandler func(context *ValidatorContext, validationData *ValidationData) error
type CallbackHandler func(context *ValidatorContext, validationData *ValidationData) []error

type BeforeTagHandler func(context *ValidatorContext, validationData *ValidationData) []error
type MiddleTagHandler func(context *ValidatorContext, validationData *ValidationData) []error
type AfterTagHandler func(context *ValidatorContext, validationData *ValidationData) []error

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidatorContext struct {
	validator *Validator
	Values    map[string]*Data
}

type ValidationData struct {
	Id             string
	Code           string
	Arguments      []interface{}
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
