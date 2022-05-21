package rule

import (
	"fmt"
	"reflect"
)

type CustomFn func(evalParam interface{}, arguments ...interface{}) (interface{}, error)

var _buildInCustomFn = map[string]CustomFn{
	// func(inValue,arr[0],arr[1])
	"in": func(evalParam interface{}, arguments ...interface{}) (interface{}, error) {
		if len(arguments) == 0 {
			return false, fmt.Errorf("no args with func `in`")
		}
		if len(arguments) == 1 {
			return false, nil
		}

		var (
			key = arguments[0]
			arr []interface{}
		)

		arr = append(arr, arguments[1:]...)
		for _, arg := range arr {
			if reflect.DeepEqual(key, arg) {
				return true, nil
			}
		}

		return false, nil
	},
}
