package jira

import (
	"context"
	"io"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableGlobalSetting(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_global_setting",
		Description: "Returns the global settings in Jira.",
		List: &plugin.ListConfig{
			Hydrate: listGlobalSettings,
		},
		Columns: commonColumns([]*plugin.Column{
			// top fields
			{
				Name:        "voting_enabled",
				Description: "Whether the ability for users to vote on issues is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "watching_enabled",
				Description: "Whether the ability for users to watch issues is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "unassigned_issues_allowed",
				Description: "Whether the ability to create unassigned issues is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "sub_tasks_enabled",
				Description: "Whether the ability to create subtasks for issues is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "issue_linking_enabled",
				Description: "Whether the ability to link issues is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "time_tracking_enabled",
				Description: "Whether the ability to track time is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "attachments_enabled",
				Description: "Whether the ability to add attachments to issues is enabled.",
				Type:        proto.ColumnType_BOOL,
			},

			// JSON fields
			{
				Name:        "time_tracking_configuration",
				Description: "The configuration of time tracking.",
				Type:        proto.ColumnType_JSON,
			},
		}),
	}
}

//// LIST FUNCTION

func listGlobalSettings(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_global_setting.listGlobalSettings", "connection_error", err)
		return nil, err
	}

	req, err := client.NewRequest("GET", "/rest/api/3/configuration", nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_global_setting.listGlobalSettings", "get_request_error", err)
		return nil, err
	}

	listGlobalSettings := new(GlobalSetting)
	res, err := client.Do(req, listGlobalSettings)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_global_setting.listGlobalSettings", "res_body", string(body))

	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_global_setting.listGlobalSettings", "api_error", err)
		return nil, err
	}

	d.StreamListItem(ctx, listGlobalSettings)
	return nil, err
}

type GlobalSetting struct {
	VotingEnabled             bool               `json:"votingEnabled"`
	WatchingEnabled           bool               `json:"watchingEnabled"`
	UnassignedIssuesAllowed   bool               `json:"unassignedIssuesAllowed"`
	SubTasksEnabled           bool               `json:"subTasksEnabled"`
	IssueLinkingEnabled       bool               `json:"issueLinkingEnabled"`
	TimeTrackingEnabled       bool               `json:"timeTrackingEnabled"`
	AttachmentsEnabled        bool               `json:"attachmentsEnabled"`
	TimeTrackingConfiguration TimeTrackingConfig `json:"timeTrackingConfiguration"`
}

type TimeTrackingConfig struct {
	WorkingHoursPerDay float64 `json:"workingHoursPerDay"`
	WorkingDaysPerWeek float64 `json:"workingDaysPerWeek"`
	TimeFormat         string  `json:"timeFormat"`
	DefaultUnit        string  `json:"defaultUnit"`
}
