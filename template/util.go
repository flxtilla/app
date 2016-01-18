package template

import (
	"reflect"

	"github.com/thrisp/flotilla/xrr"
)

func canBeNil(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	}
	return false
}

var WrongNumberArgs = xrr.NewXrror("Wrong number of args: received %d, but expected at least %d").Out

var WrongArgType = xrr.NewXrror("Argument %d has type %s -- should be %s").Out

func call(fn reflect.Value, args ...interface{}) (interface{}, error) {
	typ := fn.Type()
	numIn := typ.NumIn()
	var dddType reflect.Type
	if typ.IsVariadic() {
		if len(args) < numIn-1 {
			return nil, WrongNumberArgs(len(args), numIn-1)
		}
		dddType = typ.In(numIn - 1).Elem()
	} else {
		if len(args) != numIn {
			return nil, WrongNumberArgs(len(args), numIn)
		}
	}
	argv := make([]reflect.Value, len(args))
	for i, arg := range args {
		value := reflect.ValueOf(arg)
		// Compute the expected type. Clumsy because of variadics.
		var argType reflect.Type
		if !typ.IsVariadic() || i < numIn-1 {
			argType = typ.In(i)
		} else {
			argType = dddType
		}
		if !value.IsValid() && canBeNil(argType) {
			value = reflect.Zero(argType)
		}
		if !value.Type().AssignableTo(argType) {
			return nil, WrongArgType(i, value.Type(), argType)
		}
		argv[i] = value
	}
	result := fn.Call(argv)
	if len(result) == 2 && !result[1].IsNil() {
		return result[0].Interface(), result[1].Interface().(error)
	}
	return result[0].Interface(), nil
}

var rferrorType = reflect.TypeOf((*error)(nil)).Elem()

func goodFunc(typ reflect.Type) bool {
	switch {
	case typ.NumOut() == 1:
		return true
	case typ.NumOut() == 2 && typ.Out(1) == rferrorType:
		return true
	}
	return false
}

var NotAFunction = xrr.NewXrror("Provided (%+v, type: %T), but it is not a function").Out

var BadFunc = xrr.NewXrror("Cannot use function %q with %d results\nreturn must be 1 value, or 1 value and 1 error value").Out

func valueFunc(fn interface{}) reflect.Value {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(NotAFunction(fn, fn))
	}
	if !goodFunc(v.Type()) {
		panic(BadFunc(fn, v.Type().NumOut()))
	}
	return v
}
