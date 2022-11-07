package common

import (
	"Plow/plow/utility"
	"strings"
)

var (
	bad4db = [...]string{"'", "@", "!", ";", "#", "\"", "$", "%", "^", "&", "*", "(", ")", "/"}
)

func MakeStringDatabaseSafe(input string) string {
	output := input
	for _, v := range bad4db {
		output = strings.ReplaceAll(output, v, " ")
	}
	return output
}

func SegmentScopeCommands(blob string) []string {
	return utility.Filter(utility.Map(strings.FieldsFunc(blob, func(c rune) bool {
		return c == ';'
	}), func(input string) string {
		return strings.ReplaceAll(input, "\n", "")
	}), func(input string) bool {
		return len(strings.TrimSpace(input)) > 0
	})
}
