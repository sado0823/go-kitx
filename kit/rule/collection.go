package rule

import (
	"reflect"
	"strconv"
)

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

func reflectSelect(key string, value interface{}) (selection interface{}, ok bool) {
	vv := reflect.ValueOf(value)
	vvElem := resolvePotentialPointer(vv)

	switch vvElem.Kind() {
	case reflect.Map:
		mapKey, ok := reflectConvertTo(vv.Type().Key().Kind(), key)
		if !ok {
			return nil, false
		}

		vvElem = vv.MapIndex(reflect.ValueOf(mapKey))
		vvElem = resolvePotentialPointer(vvElem)

		if vvElem.IsValid() {
			return vvElem.Interface(), true
		}
	case reflect.Slice:
		if i, err := strconv.Atoi(key); err == nil && i >= 0 && vv.Len() > i {
			vvElem = resolvePotentialPointer(vv.Index(i))
			return vvElem.Interface(), true
		}
	case reflect.Struct:
		field := vvElem.FieldByName(key)
		if field.IsValid() {
			return field.Interface(), true
		}

		method := vv.MethodByName(key)
		if method.IsValid() {
			return method.Interface(), true
		}
	}
	return nil, false
}

func resolvePotentialPointer(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return value.Elem()
	}
	return value
}

func reflectConvertTo(k reflect.Kind, value string) (interface{}, bool) {
	switch k {
	case reflect.String:
		return value, true
	case reflect.Int:
		if i, err := strconv.Atoi(value); err == nil {
			return i, true
		}
	}
	return nil, false
}
