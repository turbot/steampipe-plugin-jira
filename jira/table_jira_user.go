package jira

import (
	"context"
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_user",
		Description: "User in the Jira cloud.",
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

			// JSON fields
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
			{
				Name:        "user_properties",
				Description: "The properties that the user have.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getUserPropertyKeys,
				Transform:   transform.From(getUserPropertiesMap),
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
	logger := plugin.Logger(ctx)
	logger.Trace("listUsers")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	req, _ := client.NewRequest("GET", "rest/api/2/users", nil)
	users := new([]jira.User)

	rsp, err := client.Do(req, users)
	if err != nil {
		logger.Error("listUsers", "Error", err)
		return nil, err
	}

	for _, user := range *users {
		d.StreamListItem(ctx, user)
	}

	return rsp, err
}

//// HYDRATE FUNCTIONS

func getUserGroups(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getUserGroups")

	user := h.Item.(jira.User)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	groups, _, err := client.User.GetGroups(user.AccountID)
	if err != nil {
		logger.Error("getUserGroups", "Error", err)
		return nil, err
	}

	return groups, nil
}

func getUserPropertyKeys(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getUserPropertyKeys")

	user := h.Item.(jira.User)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf(
		"/rest/api/3/user/properties?accountId=%s",
		user.AccountID,
	)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	userProperties := new(PropertyKeys)

	_, err = client.Do(req, userProperties)
	if err != nil {
		plugin.Logger(ctx).Error("getUserPropertyKeys", "Error", err)
		return nil, err
	}
	logger.Trace("getUserPropertyKeys", "User Properties", userProperties)
	return userProperties, nil
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

func getUserPropertiesMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	keys := d.HydrateItem.(*PropertyKeys).Keys
	userPropertiesMap := make(map[string]string)
	if keys != nil {
		for _, i := range keys {
			userPropertiesMap[i.Key] = i.Self
		}
	}
	return userPropertiesMap, nil
}

//// Custom Structs

type PropertyKeys struct {
	Keys []key `json:"keys"`
}

type key struct {
	Self string `json:"self"`
	Key  string `json:"key"`
}
