package validator

import (
	"bufio"
	"github.com/joaosoft/logger"
	"io"
	"os"
	"strings"
)

func NewValidator() *Validator {

	v := &Validator{
		tag:       constDefaultValidationTag,
		callbacks: make(map[string]callbackHandler),
		sanitize:  make([]string, 0),
		logger:    logger.NewLogDefault(constDefaultLogTag, logger.InfoLevel),
	}

	v.init()

	return v
}

func (v *Validator) newActiveHandlers() map[string]empty {
	handlers := make(map[string]empty)

	for key, _ := range v.handlersBefore {
		handlers[key] = empty{}
	}

	for key, _ := range v.handlersMiddle {
		handlers[key] = empty{}
	}

	for key, _ := range v.handlersAfter {
		handlers[key] = empty{}
	}

	return handlers
}

func (v *Validator) loadPasswords() (_ map[string]empty, err error) {
	passwords := make(map[string]empty)
	var file *os.File

	file, err = os.Open("./conf/passwords.txt")
	if err != nil {
		return passwords, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}

		passwords[strings.TrimSuffix(line, "\n")] = empty{}

		if err != nil {
			break
		}
	}
	if err != io.EOF {
		return passwords, err
	}

	return passwords, nil
}

func (v *Validator) AddBefore(name string, handler beforeTagHandler) *Validator {
	v.handlersBefore[name] = handler
	v.activeHandlers[name] = empty{}

	return v
}

func (v *Validator) AddMiddle(name string, handler middleTagHandler) *Validator {
	v.handlersMiddle[name] = handler
	v.activeHandlers[name] = empty{}

	return v
}

func (v *Validator) AddAfter(name string, handler afterTagHandler) *Validator {
	v.handlersAfter[name] = handler
	v.activeHandlers[name] = empty{}

	return v
}

func (v *Validator) SetErrorCodeHandler(handler errorCodeHandler) *Validator {
	v.errorCodeHandler = handler

	return v
}

func (v *Validator) SetValidateAll(canValidateAll bool) *Validator {
	v.canValidateAll = canValidateAll

	return v
}

func (v *Validator) SetTag(tag string) *Validator {
	v.tag = tag

	return v
}

func (v *Validator) SetSanitize(sanitize []string) *Validator {
	v.sanitize = sanitize

	return v
}

func (v *Validator) AddCallback(name string, callback callbackHandler) *Validator {
	v.callbacks[name] = callback

	return v
}

func (v *Validator) Validate(obj interface{}, args ...*argument) []error {
	return NewValidatorHandler(v, args...).handleValidation(obj)
}
