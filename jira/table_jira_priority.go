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
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_priority.listPriorities", "connection_error", err)
		return nil, err
	}

	req, err := client.NewRequest("GET", "rest/api/3/priority", nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_priority.listPriorities", "get_request_error", err)
		return nil, err
	}
	priorities := new([]jira.Priority)

	res, err := client.Do(req, priorities)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_priority.listPriorities", "res_body", string(body))

	if err != nil {
		plugin.Logger(ctx).Error("jira_priority.listPriorities", "api_error", err)
		return nil, err
	}

	for _, priority := range *priorities {
		d.StreamListItem(ctx, priority)
		// Context may get cancelled due to manual cancellation or if the limit has been reached
		if d.QueryStatus.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, err
}

//// HYDRATE FUNCTIONS

func getPriority(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_priority.getPriority", "connection_error", err)
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
		plugin.Logger(ctx).Error("jira_priority.getPriority", "get_request_error", err)
		return nil, err
	}
	result := new(jira.Priority)

	res, err := client.Do(req, result)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_priority.getPriority", "res_body", string(body))

	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_priority.getPriority", "api_error", err)
		return nil, err
	}

	return result, nil
}
