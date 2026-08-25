package util

import (
	"strings"
)

func AnyBlank(s ...string) bool {
	for _, str := range s {
		if strings.TrimSpace(str) == "" {
			return true
		}
	}
	return false
}
