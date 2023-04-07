package validator

import (
	"fmt"
	"github.com/pkg/errors"
	reflect "reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

var Validators = map[string]Validator{"len": &ValidateLen{}, "min": &ValidateMin{},
	"max": &ValidateMax{}, "in": &ValidateIn{}}

type ValidationErrors []ValidationError

type Validator interface {
	Valid(value reflect.Value, sizeMin string) error
}

func (v ValidationErrors) Error() string {
	errors := ""
	for _, err := range v {
		errors = errors + err.Err.Error()
	}
	fmt.Println(errors)
	return errors
}

func Validate(v any) error {
	var errors ValidationErrors
	typeI := reflect.TypeOf(v)
	valueI := reflect.ValueOf(v)
	if typeI.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	for i := 0; i < typeI.NumField(); i++ {

		curField := valueI.Field(i)
		var err error

		tags := ParseValidator(typeI.Field(i).Tag.Get("validate"))

		for j := 0; j < len(tags); j += 2 {

			if curField.Type().Kind() == reflect.Slice {
				for i := 0; i < curField.Len(); i++ {
					err := Validators[tags[j]].Valid(curField.Index(i), tags[j+1])
					if err != nil {
						errors = append(errors, ValidationError{err})
					}
				}
			} else if len(tags)%2 == 0 {
				err = Validators[tags[j]].Valid(curField, tags[j+1])
				if err != nil {
					errors = append(errors, ValidationError{err})
				}
			}

		}
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ParseValidator(string2 string) []string {
	curTags := strings.Split(string2, ";")
	tags := make([]string, 0)
	for _, val := range curTags {
		tags = append(tags, strings.Split(val, ":")...)
	}
	return tags
}

type ValidateLen struct {
}

func (*ValidateLen) Valid(value reflect.Value, sizeM string) error {
	sizeNew, err := strconv.Atoi(sizeM)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}
	if value.Type().Kind() != reflect.String {
		return ErrInvalidValidatorSyntax
	}
	if len([]rune(value.String())) != sizeNew {
		return ErrValidateForUnexportedFields
	}
	return nil
}

type ValidateMin struct {
}

func (*ValidateMin) Valid(value reflect.Value, sizeMin string) error {
	size, err := strconv.Atoi(sizeMin)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}
	switch value.Type().Kind() {
	case reflect.String:
		if len(value.String()) < size {
			return ErrValidateForUnexportedFields
		}
	case reflect.Int:
		if int(value.Int()) < size {
			return ErrValidateForUnexportedFields
		}
	}
	return nil
}

type ValidateMax struct {
}

func (*ValidateMax) Valid(value reflect.Value, sizeMax string) error {
	size, err := strconv.Atoi(sizeMax)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}
	switch value.Type().Kind() {
	case reflect.String:
		if len(value.String()) > size {
			return ErrValidateForUnexportedFields
		}
	case reflect.Int:
		if int(value.Int()) > size {
			return ErrValidateForUnexportedFields
		}
	}
	return nil
}

type ValidateIn struct {
}

func (*ValidateIn) Valid(value reflect.Value, tag string) error {
	arr := strings.Split(tag, ",")
	var err error
	switch value.Type().Kind() {
	case reflect.Int:
		arrNew, err := ConvertToArrayInt(arr)
		if err != nil {
			return err
		}
		val := int(value.Int())
		err = Contains(arrNew, val)
		if err != nil {
			return err
		}
	case reflect.String:
		val := value.String()
		if val == "" {
			return ErrInvalidValidatorSyntax
		}
		err = Contains(arr, val)
		if err != nil {
			return err
		}
	default:
		return ErrInvalidValidatorSyntax
	}
	return nil
}

func Contains[T comparable](t []T, needle T) error {
	for _, v := range t {
		if v == needle {
			return nil
		}
	}
	return ErrInvalidValidatorSyntax
}

func ConvertToArrayInt(t []string) ([]int, error) {
	var t2 = make([]int, len(t))
	for idx, i := range t {
		j, err := strconv.Atoi(i)
		if err != nil {
			return nil, ErrInvalidValidatorSyntax
		}
		t2[idx] = j
	}
	return t2, nil
}
