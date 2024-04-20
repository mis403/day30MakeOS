package msgo

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"sync"
)

type StructValidator interface {
	ValidateStruct(any) error //结构体验证
	Engine() any
}

var Validator StructValidator = &defaultValidator{}

type defaultValidator struct {
	one      sync.Once
	validate *validator.Validate
}

func (d *defaultValidator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}
	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return d.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return d.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}
func (d *defaultValidator) validateStruct(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}

// 多线程环境下，只有第一次调用时才会执行其中的代码,并初始化d.validate，并且以后的调用将不再执行初始化过程
func (d *defaultValidator) lazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}
