package utility

import (
	"fmt"
	"strings"
)

func IsStringEmpty(input *string) bool {
	if input != nil {
		if len(strings.TrimSpace(*input)) > 0 {
			return false
		}
	}
	return true
}

func LeadingRepeatSprintf(n int, repeat string, format string, values ...interface{}) string {
	return strings.Repeat(repeat, n) + fmt.Sprintf(format, values...)
}

func TabbedPrintln(n int, value string) {
	fmt.Println(LeadingRepeatSprintf(n, "\t", "%s", value))
}

func TabbedPrintlnf(n int, format string, values ...interface{}) {
	TabbedPrintln(n, fmt.Sprintf(format, values...))
}
