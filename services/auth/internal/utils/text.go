package utils

import (
	"regexp"
	"strings"
)

func ToSnakeCase(str string) string {
	pattern := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := pattern.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}