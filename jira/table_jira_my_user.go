package jira

import (
	"context"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableMyUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_my_user",
		Description:      "Details of the your own jira user.",
		List: &plugin.ListConfig{
			Hydrate: getMyUser,
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
				Description: "The email address for the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "active",
				Description: "Indicates if user is active.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "self",
				Description: "The URL of the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "locale",
				Description: "The locale of the user. Depending on the user’s privacy setting, this may be returned as null.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "time_zone",
				Description: "The time zone specified in the user's profile. Depending on the user’s privacy setting, this may be returned as null.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "avatar_urls",
				Description: "The avatars of the user.",
				Type:        proto.ColumnType_JSON,
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

func getMyUser(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	req, _ := client.NewRequest("GET", "/rest/api/2/myself", nil)

	myself := new(jira.User)

	_, err = client.Do(req, myself)
	if err != nil {
		return nil, err
	}

	d.StreamListItem(ctx, myself)

	return nil, err
}
