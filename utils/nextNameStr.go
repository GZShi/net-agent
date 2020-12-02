package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

// NextNameStr 出现重名时，在后面加上数字后缀
func NextNameStr(str string) string {
	replaced := false
	re := regexp.MustCompile(`_(\d*)$`)
	nextStr := re.ReplaceAllStringFunc(str, func(s string) string {
		replaced = true
		if s == "_" {
			return "_1"
		}
		lenS := len(s) - 1
		index, err := strconv.Atoi(s[1:])
		if err != nil {
			return "_1"
		}

		index++
		n := fmt.Sprintf("%v", index)
		for len(n) < lenS {
			n = "0" + n
		}
		return "_" + n
	})

	if !replaced {
		return str + "_1"
	}
	return nextStr
}
