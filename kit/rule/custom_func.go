package rule

import (
	"fmt"
	"reflect"
)

type CustomFn func(arguments ...interface{}) (interface{}, error)

var _buildInCustomFn = map[string]CustomFn{
	// func(inValue,arr[0],arr[1])
	"in": func(arguments ...interface{}) (interface{}, error) {
		logger.Println("build func [in], len args=", len(arguments))
		logger.Println(arguments...)

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
		logger.Printf("key=%v, key_type=%T \n", key, key)
		for _, arg := range arr {
			logger.Printf("arg=%v, arg_type=%T \n", arg, arg)
			if reflect.DeepEqual(key, arg) {
				return true, nil
			}
		}

		return false, nil
	},
}
