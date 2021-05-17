package util

import (
	"reflect"
)

// analog of java generics
//
func ItemExists(arrayType interface{}, item interface{}) bool {
	arr := reflect.ValueOf(arrayType)

	if arr.Kind() != reflect.Array {
		panic("Invalid data-type")
	}

	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

func IsArray(val interface{}) bool {
	return GetKind(val) == reflect.Array
}

func IsString(val interface{}) bool {
	return GetKind(val) == reflect.String
}

func IsSlice(val interface{}) bool {
	return GetKind(val) == reflect.Slice
}

func GetKind(val interface{}) reflect.Kind {
	return reflect.ValueOf(val).Kind()
}
