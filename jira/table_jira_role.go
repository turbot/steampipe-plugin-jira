package jira

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableProjectRole(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_project_role",
		Description:      "Roles are a flexible way to associate users and/or groups with particular projects.",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getProjectRole,
		},
		List: &plugin.ListConfig{
			Hydrate: listProjectRoles,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The ID of the project role.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The name of the project role.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name: "self",
				Type: proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "A description of the project role.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name: "actors",
				Type: proto.ColumnType_JSON,
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		},
	}
}

//// LIST FUNCTION

func listProjectRoles(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	roles, _, err := client.Role.GetList()
	if err != nil {
		return nil, err
	}

	for _, role := range *roles {
		d.StreamListItem(ctx, role)
	}

	return nil, err
}

//// HYDRATE FUNCTION

func getProjectRole(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	roleId := d.KeyColumnQuals["id"].GetInt64Value()

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	if roleId == 0 {
		return nil, nil
	}

	role, _, err := client.Role.Get(int(roleId))
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return role, err
}
