package util

import "strings"

func LogSanitize(msg string) string {
	escapedMsg := strings.Replace(msg, "\n", "", -1)
	escapedMsg = strings.Replace(escapedMsg, "\r", "", -1)
	return strings.TrimSpace(escapedMsg)
}
