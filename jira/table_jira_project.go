package jira

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableProject(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_project",
		Description:      "Jira Project",
		DefaultTransform: transform.FromCamel(),
		List: &plugin.ListConfig{
			Hydrate: listProjects,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The account ID of the user, which uniquely identifies the user across all Atlassian products. For example, 5b10ac8d82e05b22cc7d4ef5.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The email address for the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "key",
				Description: "The email address for the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name: "self",
				Type: proto.ColumnType_STRING,
			},
			{
				Name: "project_type_key",
				Type: proto.ColumnType_STRING,
			},
			{
				Name: "project_category",
				Type: proto.ColumnType_JSON,
			},
			{
				Name: "issue_types",
				Type: proto.ColumnType_JSON,
			},
			{
				Name: "avatar_urls",
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

func listProjects(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	projects, _, err := client.Project.GetList()
	if err != nil {
		return nil, err
	}

	for _, project := range *projects {
		d.StreamListItem(ctx, project)
	}

	return nil, err
}
