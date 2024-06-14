package jira

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/memoize"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Login ID would be same per connection.
// API key and API Token is specific to a User
// A user can have multiple organizations, workspaces, boards, etc...
func commonColumns(c []*plugin.Column) []*plugin.Column {
	return append([]*plugin.Column{
		{
			Name:        "login_id",
			Type:        proto.ColumnType_STRING,
			Description: "The unique identifier of the user login.",
			Hydrate:     getLoginId,
			Transform:   transform.FromValue(),
		},
	}, c...)
}

// if the caching is required other than per connection, build a cache key for the call and use it in Memoize.
var getLoginIdMemoized = plugin.HydrateFunc(getLoginIdUncached).Memoize(memoize.WithCacheKeyFunction(getLoginIdCacheKey))

// declare a wrapper hydrate function to call the memoized function
// - this is required when a memoized function is used for a column definition
func getLoginId(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return getLoginIdMemoized(ctx, d, h)
}

// Build a cache key for the call to getLoginIdCacheKey.
func getLoginIdCacheKey(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	key := "getLoginId"
	return key, nil
}

func getLoginIdUncached(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Create client
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("getLoginIdUncached", "connection_error", err)
		return nil, err
	}

	currentSession, _, err := client.User.GetSelf()
	if err != nil {
		plugin.Logger(ctx).Error("getLoginIdUncached", "api_error", err)
		return nil, err
	}

	return currentSession.AccountID, nil
}
