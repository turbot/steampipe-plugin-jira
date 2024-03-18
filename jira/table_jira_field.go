package jira

import (
	"context"
	"io"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableField(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_field",
		Description: "Details of the fields.",

		List: &plugin.ListConfig{
			Hydrate: listFields,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the field.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The ID of the field.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "key",
				Description: "The key of the field.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "custom",
				Description: "The flag to indicate if the field is custom.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("Custom"),
			},
			{
				Name:        "searchable",
				Description: "The flag to indicate if the field is searchable.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("Searchable"),
			},
			{
				Name:        "schema.type",
				Description: "Data type of the field.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Schema.Type"),
			},
			{
				Name:        "schema",
				Description: "Data type of the field.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Schema"),
			},

			// Steampipe standard colu@mns
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

func listFields(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_field.listFields", "connection_error", err)
		return nil, err
	}

	req, err := client.NewRequest("GET", "rest/api/3/field", nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_field.listFields", "get_request_error", err)
		return nil, err
	}
	fields := new([]jira.Field)

	res, err := client.Do(req, fields)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_field.listFields", "res_body", string(body))

	if err != nil {
		plugin.Logger(ctx).Error("jira_field.listFields", "api_error", err)
		return nil, err
	}

	for _, priority := range *fields {
		d.StreamListItem(ctx, priority)
		// Context may get cancelled due to manual cancellation or if the limit has been reached
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, err
}

//// HYDRATE FUNCTIONS
