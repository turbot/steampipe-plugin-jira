package jira

import (
	"context"
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableIssueWorklog(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_issue_worklog",
		Description: "Jira worklog is a feature within the Jira software that allows users to record the amount of time they have spent working on various tasks or issues.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AnyColumn([]string{"issue_id", "id"}),
			Hydrate:    getIssueWorklog,
		},
		List: &plugin.ListConfig{
			ParentHydrate: listIssues,
			Hydrate:       listIssueWorklogs,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "issue_id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "A unique identifier for the worklog entry.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "issue_id",
				Description: "The ID of the issue.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the worklogs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "comment",
				Description: "Any comments or descriptions added to the worklog entry.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "started",
				Description: "The date and time when the worklog activity started.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Started").Transform(convertJiraTime),
			},
			{
				Name:        "created",
				Description: "The date and time when the worklog entry was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Created").Transform(convertJiraTime),
			},
			{
				Name:        "updated",
				Description: "The date and time when the worklog entry was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Updated").Transform(convertJiraTime),
			},
			{
				Name:        "time_spent",
				Description: "The duration of time logged for the task, often in hours or minutes.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "time_spent_seconds",
				Description: "The duration of time logged in seconds.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "properties",
				Description: "The properties of each worklog.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "author",
				Description: "Information about the user who created the worklog entry, often including their username, display name, and user account details.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "update_author",
				Description: "Details of the user who last updated the worklog entry, similar to the author information.",
				Type:        proto.ColumnType_JSON,
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
		}),
	}
}

type WorklogDetails struct {
	jira.WorklogRecord
	IssueId string
}

//// LIST FUNCTION

func listIssueWorklogs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if h.Item == nil {
		return nil, nil
	}
	issueinfo := h.Item.(IssueInfo)
	issueId := d.EqualsQualString("issue_id")

	// Minize the API call for given issue ID.
	if issueId != "" && issueId != issueinfo.ID {
		return nil, nil
	}

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_worklog.listIssueWorklogs", "connection_error", err)
		return nil, err
	}

	last := 0

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var limit int = 5000
	if d.QueryContext.Limit != nil {
		if *queryLimit < 5000 {
			limit = int(*queryLimit)
		}
	}

	for {
		apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/worklog?startAt=%d&maxResults=%d&expand=properties", issueinfo.ID, last, limit)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_worklog.listIssueWorklogs", "get_request_error", err)
			return nil, err
		}

		w := new(jira.Worklog)
		_, err = client.Do(req, w)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_worklog.listIssueWorklogs", "api_error", err)
			return nil, err
		}

		for _, c := range w.Worklogs {
			d.StreamListItem(ctx, WorklogDetails{c, issueinfo.ID})

			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		last = w.StartAt + len(w.Worklogs)
		if last >= w.Total {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getIssueWorklog(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	issueId := d.EqualsQualString("issue_id")
	id := d.EqualsQualString("id")

	if issueId == "" || id == "" {
		return nil, nil
	}

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_worklog.getIssueWorklog", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/worklog/%s?expand=properties", issueId, id)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_worklog.getIssueWorklog", "get_request_error", err)
		return nil, err
	}

	res := new(jira.WorklogRecord)
	_, err = client.Do(req, res)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_issue_worklog.getIssueWorklog", "api_error", err)
		return nil, err
	}

	return WorklogDetails{*res, issueId}, nil
}
