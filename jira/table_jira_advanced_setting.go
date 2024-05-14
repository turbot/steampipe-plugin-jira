package jira

import (
	"context"
	"fmt"
	"io"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableAdvancedSetting(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_advanced_setting",
		Description: "The application properties that are accessible on the Advanced Settings page.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getAdvancedSettingProperty,
		},
		List: &plugin.ListConfig{
			Hydrate: listAdvancedSettings,
		},
		Columns: commonColumns([]*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the application property.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The name of the application property.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the application property.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Description"),
			},

			// other important fields
			{
				Name:        "key",
				Description: "The key of the application property.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The data type of the application property.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "value",
				Description: "The new value.",
				Type:        proto.ColumnType_STRING,
			},

			// JSON fields
			{
				Name:        "allowed_values",
				Description: "The allowed values, if applicable.",
				Type:        proto.ColumnType_JSON,
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		}),
	}
}

//// LIST FUNCTION

func listAdvancedSettings(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_advanced_setting.listAdvancedSettings", "connection_error", err)
		return nil, err
	}

	req, err := client.NewRequest("GET", "/rest/api/3/application-properties/advanced-settings", nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_advanced_setting.listAdvancedSettings", "get_request_error", err)
		return nil, err
	}

	listAdvancedSettings := new([]AdvancedApplicationProperty)
	res, err := client.Do(req, listAdvancedSettings)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_advanced_setting.listAdvancedSettings", "res_body", string(body))

	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_advanced_setting.listAdvancedSettings", "api_error", err)
		return nil, err
	}

	for _, listAdvancedSettings := range *listAdvancedSettings {
		d.StreamListItem(ctx, listAdvancedSettings)
	}
	return nil, err
}

//// HYDRATE FUNCTIONS

func getAdvancedSettingProperty(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	ID := d.EqualsQuals["id"].GetStringValue()

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_advanced_setting.getAdvancedSettingProperty", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/application-properties?key=%s", ID)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_advanced_setting.getAdvancedSettingProperty", "get_request_error", err)
		return nil, err
	}

	result := new(AdvancedApplicationProperty)

	res, err := client.Do(req, result)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_advanced_setting.getAdvancedSettingProperty", "res_body", string(body))

	if err != nil {
		if isBadRequestError(err) || isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_advanced_setting.getAdvancedSettingProperty", "api_error", err)
		return nil, err
	}

	return result, nil
}

type AdvancedApplicationProperty struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	Value         string   `json:"value"`
	Name          string   `json:"name"`
	Description   string   `json:"desc"`
	Type          string   `json:"type"`
	AllowedValues []string `json:"allowedValues"`
}
