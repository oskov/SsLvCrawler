package util

import (
	"fmt"
	"regexp"
	"time"
)

func FilterChars(s string, pattern string) (result string) {
	reg := regexp.MustCompile(pattern)
	result = string(reg.ReplaceAll([]byte(s), []byte("")))
	return result
}

func CurrentDateTime() (result string) {
	t := time.Now()
	result = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return result
}
