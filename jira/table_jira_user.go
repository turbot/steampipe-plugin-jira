package jira

import (
	"context"
	"fmt"
	"io"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v4/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
)

//// TABLE DEFINITION

func tableUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_user",
		Description: "User in the Jira cloud.",
		List: &plugin.ListConfig{
			Hydrate: listUsers,
		},
		HydrateConfig: []plugin.HydrateConfig{
			{
				// Limit concurrency to avoid a 429 too many requests error
				Func:           getUserGroups,
				MaxConcurrency: 50,
		},
		Columns: []*plugin.Column{
			{
				Name:        "display_name",
				Description: "The display name of the user. Depending on the user's privacy setting, this may return an alternative value.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "account_id",
				Description: "The account ID of the user, which uniquely identifies the user across all Atlassian products. For example, 5b10ac8d82e05b22cc7d4ef5.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "email_address",
				Description: "The email address of the user. Depending on the user's privacy setting, this may be returned as null.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "account_type",
				Description: "The user account type. Can take the following values: atlassian, app, customer and unknown.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "active",
				Description: "Indicates if user is active.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("Active"),
			},
			{
				Name:        "self",
				Description: "The URL of the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "avatar_urls",
				Description: "The avatars of the user.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "group_names",
				Description: "The groups that the user belongs to.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getUserGroups,
				Transform:   transform.From(groupNames),
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DisplayName"),
			},
		},
	}
}

//// LIST FUNCTION

func listUsers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_user.listUsers", "connection_error", err)
		return nil, err
	}

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var maxResults int = 1000
	if d.QueryContext.Limit != nil {
		if *queryLimit < 1000 {
			maxResults = int(*queryLimit)
		}
	}

	last := 0
	for {
		apiEndpoint := fmt.Sprintf("rest/api/2/users/search?startAt=%d&maxResults=%d", last, maxResults)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_user.listUsers", "get_request_error", err)
			return nil, err
		}

		users := new([]jira.User)
		res, err := client.Do(req, users)
		body, _ := io.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_user.listUsers", "res_body", string(body))

		if err != nil {
			plugin.Logger(ctx).Error("jira_user.listUsers", "api_error", err)
			return nil, err
		}

		for _, user := range *users {
			d.StreamListItem(ctx, user)
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.QueryStatus.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		// evaluate paging start value for next iteration
		last = last + len(*users)

		// API doesn't gives paging parameters in the response,
		// therefore using output length to quit paging
		if len(*users) < 1000 {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTIONS

func getUserGroups(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	user := h.Item.(jira.User)

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_user.getUserGroups", "connection_error", err)
		return nil, err
	}

	groups, _, err := client.User.GetGroups(user.AccountID)

	if err != nil {
		plugin.Logger(ctx).Error("jira_user.getUserGroups", "api_error", err)
		return nil, err
	}

	return groups, nil
}

//// TRANSFORM FUNCTION

func groupNames(_ context.Context, d *transform.TransformData) (interface{}, error) {
	userGroups := d.HydrateItem.(*[]jira.UserGroup)
	var groupNames []string
	for _, group := range *userGroups {
		groupNames = append(groupNames, group.Name)
	}
	return groupNames, nil
}
