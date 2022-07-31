package reflectx

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// PathSelect select key from path by reflectx.
//
// Support for hierarchical connections of point; Support data type `Struct`(Should be Exported Field), `Map`, `Slice` , `Pointer`.
//
// example: `a.b.c` Or with slice `a.b.2.c`, can see detail from unit test file `reflect_test.go`
func PathSelect(ctx context.Context, path string, value interface{}) (selection interface{}, ok bool) {

	selection = value
	keys := strings.Split(path, ".")

	for _, key := range keys {
		selection, ok = reflectSelect(key, selection)
		if !ok {
			return selection, false
		}
	}

	return selection, true
}

func reflectSelect(key string, value interface{}) (selection interface{}, ok bool) {
	if len(key) == 0 {
		return nil, false
	}

	vv := reflect.ValueOf(value)
	vvElem := resolvePotentialPointer(vv)

	vType := reflect.TypeOf(value)
	if vType == nil {
		return nil, false
	}
	vTypeElem := resolvePotentialPointerType(vType)

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
		// by struct field
		field := vvElem.FieldByName(key)
		if field.IsValid() {
			// unexported field
			firstCharacter := []rune(key)[0]
			if unicode.ToUpper(firstCharacter) != firstCharacter {
				return nil, false
			}
			return field.Interface(), true
		}

		// by json Tag
		v, ok := reflectTagValue(vvElem, vTypeElem, "json", key)
		if ok {
			return v, true
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

func resolvePotentialPointerType(value reflect.Type) reflect.Type {
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

func reflectTagValue(value reflect.Value, vType reflect.Type, tagName string, key string) (interface{}, bool) {

	// by json Tag
	vvElem := resolvePotentialPointer(value)
	vTypeElem := resolvePotentialPointerType(vType)

	for i := 0; i < vTypeElem.NumField(); i++ {
		fieldT := vTypeElem.Field(i)
		fieldV := vvElem.FieldByName(fieldT.Name)
		lookup, ok := fieldT.Tag.Lookup(tagName)
		if ok && strings.Replace(lookup, ",omitempty", "", -1) == key && fieldV.IsValid() && fieldT.IsExported() {
			return fieldV.Interface(), true
		}

		fieldVElem := resolvePotentialPointer(fieldV)
		if fieldVElem.Kind() == reflect.Struct {
			v, ok := reflectTagValue(fieldVElem, fieldT.Type, tagName, key)
			if ok {
				return v, true
			}
		}
	}

	return nil, false
}
