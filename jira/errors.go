package jira

import (
	"context"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
)

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "404")
}

func isBadRequestError(err error) bool {
	return strings.Contains(err.Error(), "400")
}

func shouldRetryError(retryErrors []string) plugin.ErrorPredicateWithContext {
	return func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData, err error) bool {

		if strings.Contains(err.Error(), "429") {
			plugin.Logger(ctx).Warn("Received Rate Limit Error")
			return true
		}
		return false
	}
}
