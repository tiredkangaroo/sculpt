package sculpt

import "strings"

func replaceAllFunc(s, old string, replaceFunc func() string) string {
	var result strings.Builder
	var i, start int
	for i < len(s) {
		index := strings.Index(s[i:], old)
		if index == -1 {
			break
		}
		index += i
		result.WriteString(s[start:index])
		result.WriteString(replaceFunc())
		i = index + len(old)
		start = i
	}
	result.WriteString(s[start:])
	return result.String()
}
