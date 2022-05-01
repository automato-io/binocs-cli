package util

import (
	"fmt"
	"math"
	"regexp"
	"time"
)

// Ellipsis produces shortened utf-8-safe version of a string
func Ellipsis(s string, maxLen int) string {
	var l = len([]rune(s))
	if l <= maxLen {
		return s
	}
	tt := int(math.Round(float64(maxLen-3) * 0.68))
	lt := maxLen - tt
	return UtfSubstr(s, 0, tt-2) + "..." + UtfSubstr(s, l-lt+1, lt)
}

// OutputDurationWithDays extends time.Duration with number of elapsed days
func OutputDurationWithDays(d string) string {
	parsed, err := time.ParseDuration(d)
	if err != nil {
		return d
	}
	if parsed.Hours() > 48 {
		days := math.Floor(parsed.Hours() / 24)
		hours := math.Floor(parsed.Hours() - days*24)
		re1 := regexp.MustCompile(`([0-9]+)h`)
		rest := re1.ReplaceAllString(d, fmt.Sprintf("%.0f", hours)+"h")
		re2 := regexp.MustCompile(`([0-9]+)s`)
		rest = re2.ReplaceAllString(rest, "")
		return fmt.Sprintf("%.0f days %s", days, rest)
	}
	return d
}

// UtfSubstr produces utf-8-safe substr
func UtfSubstr(input string, start int, length int) string {
	asRunes := []rune(input)
	if start >= len(asRunes) {
		return ""
	}
	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}
	return string(asRunes[start : start+length])
}

func StringInSlice(needle string, haystack []string) bool {
	for _, h := range haystack {
		if needle == h {
			return true
		}
	}
	return false
}
