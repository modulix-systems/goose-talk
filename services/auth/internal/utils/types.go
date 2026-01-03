package utils

import (
	"encoding/json"
	"reflect"
	"strconv"
)

type BoolString bool

func (value BoolString) MarshalBinary() ([]byte, error) {
	return []byte(strconv.FormatBool(bool(value))), nil
}

func (value *BoolString) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	parsed, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*value = BoolString(parsed)

	return nil
}

func IsScalarType(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	nonScalarTypes := []reflect.Kind{reflect.Struct, reflect.Slice, reflect.Map}
	for _, nonScalarType := range nonScalarTypes {
		if t.Kind() == nonScalarType {
			return false
		}
	}
	return true
}