package jira

import "strings"

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "404")
}

func isBadRequestError(err error) bool {
	return strings.Contains(err.Error(), "400")
}
