package utils

import "reflect"

func MapStruct[T any](src any, dst T) T {
	if reflect.TypeOf(dst).Kind() != reflect.Pointer {
		panic("Destination should be a pointer")
	}

	dstVal := reflect.Indirect(reflect.ValueOf(dst))
	srcVal := reflect.Indirect(reflect.ValueOf(src))

	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Type().Field(i)

		if srcFieldVal := srcVal.FieldByName(dstField.Name); srcFieldVal.IsValid() {
			dstVal.Field(i).Set(srcFieldVal)
		}
	}
	
	return dst
}
