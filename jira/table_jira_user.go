package jira

import (
	"context"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_user",
		Description:      "User in the Jira cloud.",
		DefaultTransform: transform.FromCamel(),
		List: &plugin.ListConfig{
			Hydrate: listUsers,
		},
		Columns: []*plugin.Column{
			{
				Name:        "display_name",
				Description: "The display name of the user. Depending on the user’s privacy setting, this may return an alternative value.",
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
				Description: "The email address of the user. Depending on the user’s privacy setting, this may be returned as null.",
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
				Name:        "groups",
				Description: "The groups that the user belongs to.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getUserGroups,
				Transform:   transform.FromValue(),
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
		return nil, err
	}

	req, _ := client.NewRequest("GET", "rest/api/2/users", nil)

	users := new([]jira.User)

	rsp, err := client.Do(req, users)
	if err != nil {
		return nil, err
	}

	for _, user := range *users {
		d.StreamListItem(ctx, user)
	}

	return rsp, err
}

//// HYDRATE FUNCTIONS

func getUserGroups(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAdUser")
	user := h.Item.(jira.User)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	groups, _, err := client.User.GetGroups(user.AccountID)
	if err != nil {
		return nil, err
	}

	return groups, nil
}
