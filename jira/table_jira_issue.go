package jira

import (
	"context"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableIssue(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_issue",
		Description:      "Jira Issue",
		DefaultTransform: transform.FromCamel(),
		List: &plugin.ListConfig{
			KeyColumns: plugin.SingleColumn("project_key"),
			Hydrate:    listIssues,
		},
		Columns: []*plugin.Column{
			{
				Name:        "project_key",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Key"),
			},
			{
				Name:        "project_name",
				Description: "Name of the project to that issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Name"),
			},
			{
				Name:        "id",
				Description: "Issue unique identifier.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "self",
				Description: "A friendly name that identifies the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "key",
				Description: "A friendly name that identifies the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "Description of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Description"),
			},
			{
				Name:        "created",
				Description: "Time when the issue was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Created").Transform(convertTimestamp),
			},
			{
				Name:        "priority",
				Description: "Priority assigned to the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Priority.Name"),
			},
			{
				Name:        "labels",
				Description: "A list of labels applied to the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Labels"),
			},
			{
				Name:        "assignee",
				Description: "Details of the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Assignee"),
			},
			{
				Name:        "creator",
				Description: "Details of the user/application that created the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Creator"),
			},
			{
				Name:        "fields",
				Description: "Json object containing important subfields info the issue.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "components",
				Description: "Components are subsections of a project. They are used to group issues within a project into smaller parts.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Components"),
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
		},
	}
}

//// LIST FUNCTION

func listIssues(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectName := d.KeyColumnQuals["project_key"].GetStringValue()
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	var issues []jira.Issue
	for {
		opt := &jira.SearchOptions{
			MaxResults: 1000, // Max results can go up to 1000
			StartAt:    last,
		}

		chunk, resp, err := client.Issue.Search("project = "+projectName, opt)
		if err != nil {
			return nil, err
		}

		total := resp.Total
		if issues == nil {
			issues = make([]jira.Issue, 0, total)
		}

		for _, issue := range chunk {
			d.StreamListItem(ctx, issue)
		}
		last = resp.StartAt + len(chunk)
		if last >= total {
			return nil, nil
		}
	}

}
