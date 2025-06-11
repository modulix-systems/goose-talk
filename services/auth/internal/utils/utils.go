package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

func FindRootPath() string {
	dirPath, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dirPath, "go.mod")); err == nil {
			return dirPath
		}
		parentDirPath := filepath.Dir(dirPath)
		if parentDirPath == dirPath {
			break
		}
		dirPath = parentDirPath
	}
	return ""
}

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchAllCap.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
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
