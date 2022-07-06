package rule

import (
	"reflect"
)

func typeEqual(x, y interface{}) (xT, yT reflect.Type, ok bool) {
	xT = reflect.TypeOf(x)
	yT = reflect.TypeOf(y)

	return xT, yT, xT.Kind() == yT.Kind()
}

func isString(value interface{}) bool {
	switch value.(type) {
	case string:
		return true
	}
	return false
}

func convertToFloat(o interface{}) (float64, bool) {
	if i, ok := o.(float64); ok {
		return i, true
	}
	v := reflect.ValueOf(o)
	for o != nil && v.Kind() == reflect.Ptr {
		v = v.Elem()
		if !v.IsValid() {
			return 0, false
		}
		o = v.Interface()
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	}

	return 0, false
}