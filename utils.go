package flotilla

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/thrisp/flotilla/xrr"
)

var (
	rferrorType = reflect.TypeOf((*error)(nil)).Elem()

	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

func existsIn(s string, l []string) bool {
	for _, x := range l {
		if s == x {
			return true
		}
	}
	return false
}

func doAdd(s string, ss []string) []string {
	if isAppendable(s, ss) {
		ss = append(ss, s)
	}
	return ss
}

func isAppendable(s string, ss []string) bool {
	for _, x := range ss {
		if x == s {
			return false
		}
	}
	return true
}

func isFunc(fn interface{}) bool {
	return reflect.ValueOf(fn).Kind() == reflect.Func
}

func equalFunc(a, b interface{}) bool {
	if !isFunc(a) || !isFunc(b) {
		panic("flotilla : funcEqual -- type error!")
	}
	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()
	return av.InterfaceData() == bv.InterfaceData()
}

func valueFunc(fn interface{}) reflect.Value {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(xrr.NewError("flotilla : Provided (%+v, type: %T), but it is not a function", fn, fn))
	}
	if !goodFunc(v.Type()) {
		panic(xrr.NewError("flotilla: Cannot use function %q with %d results\nreturn must be 1 value, or 1 value and 1 error value", fn, v.Type().NumOut()))
	}
	return v
}

func goodFunc(typ reflect.Type) bool {
	switch {
	case typ.NumOut() == 1:
		return true
	case typ.NumOut() == 2 && typ.Out(1) == rferrorType:
		return true
	}
	return false
}

func canBeNil(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	}
	return false
}

func call(fn reflect.Value, args ...interface{}) (interface{}, error) {
	typ := fn.Type()
	numIn := typ.NumIn()
	var dddType reflect.Type
	if typ.IsVariadic() {
		if len(args) < numIn-1 {
			return nil, xrr.NewError("flotilla : wrong number of args -- got %d want at least %d", len(args), numIn-1)
		}
		dddType = typ.In(numIn - 1).Elem()
	} else {
		if len(args) != numIn {
			return nil, xrr.NewError("flotilla : wrong number of args -- got %d want %d", len(args), numIn)
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
			return nil, xrr.NewError("flotilla : arg %d has type %s -- should be %s", i, value.Type(), argType)
		}
		argv[i] = value
	}
	result := fn.Call(argv)
	if len(result) == 2 && !result[1].IsNil() {
		return result[0].Interface(), result[1].Interface().(error)
	}
	return result[0].Interface(), nil
}

func dropTrailing(path string, trailing string) string {
	if fp := strings.Split(path, "/"); fp[len(fp)-1] == trailing {
		return strings.Join(fp[0:len(fp)-1], "/")
	}
	return path
}

func validExtension(fn interface{}) error {
	if goodFunc(valueFunc(fn).Type()) {
		return nil
	}
	return xrr.NewError("function %q is not a valid Flotilla Ctx function; must be a function and return must be 1 value, or 1 value and 1 error value", fn)
}

func StatusColor(code int) (color string) {
	switch {
	case code >= 200 && code <= 299:
		color = green
	case code >= 300 && code <= 399:
		color = white
	case code >= 400 && code <= 499:
		color = yellow
	default:
		color = red
	}
	return color
}

func MethodColor(method string) (color string) {
	switch {
	case method == "GET":
		color = blue
	case method == "POST":
		color = cyan
	case method == "PUT":
		color = yellow
	case method == "DELETE":
		color = red
	case method == "PATCH":
		color = green
	case method == "HEAD":
		color = magenta
	case method == "OPTIONS":
		color = white
	}
	return color
}

func LogFmt(c *ctx) string {
	st := c.Result.RStatus
	md := c.Result.RMethod
	return fmt.Sprintf("%v |%s %3d %s| %12v | %s |%s %s %-7s %s",
		c.Result.RStop.Format("2006/01/02 - 15:04:05"),
		StatusColor(st), st, reset,
		c.Result.RLatency,
		c.Result.RRequester,
		MethodColor(md), reset, md,
		c.Result.RPath,
	)
}
