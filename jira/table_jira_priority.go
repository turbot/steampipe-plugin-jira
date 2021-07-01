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

func tablePriority(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_priority",
		Description: "Details of the issue priority.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getPriority,
		},
		List: &plugin.ListConfig{
			Hydrate: listPriorities,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the issue priority.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The ID of the issue priority.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "description",
				Description: "The description of the issue priority.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the issue priority.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "icon_url",
				Description: "The URL of the icon for the issue priority.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("IconURL"),
			},
			{
				Name:        "status_color",
				Description: "The color used to indicate the issue priority.",
				Type:        proto.ColumnType_STRING,
			},

			// Steampipe standard columns
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

func listPriorities(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("listPriorities")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	req, _ := client.NewRequest("GET", "rest/api/3/priority", nil)
	priorities := new([]jira.Priority)

	_, err = client.Do(req, priorities)
	if err != nil {
		plugin.Logger(ctx).Error("listPriorities", "Error", err)
		return nil, err
	}

	for _, priority := range *priorities {
		d.StreamListItem(ctx, priority)
	}

	return nil, err
}

//// HYDRATE FUNCTIONS

func getPriority(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getPriority")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}
	priorityId := d.KeyColumnQuals["id"].GetStringValue()

	// Return nil, if no input provided
	if priorityId == "" {
		return nil, nil
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/priority/%s", priorityId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}
	result := new(jira.Priority)

	_, err = client.Do(req, result)
	if err != nil {
		plugin.Logger(ctx).Error("getPriority", "Error", err)
		return nil, err
	}

	return result, nil
}
