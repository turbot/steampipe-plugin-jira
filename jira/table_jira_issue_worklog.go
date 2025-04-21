package jira

import (
	"context"
	"fmt"
	"net/url"
	"time"

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
			KeyColumns: plugin.AllColumns([]string{"issue_id", "id"}),
			Hydrate:    getIssueWorklog,
		},
		List: &plugin.ListConfig{
			Hydrate: listWorklogs,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "issue_id", Require: plugin.Optional},
				{Name: "updated", Require: plugin.Optional, Operators: []string{">", ">="}},
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

func listWorklogs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if d.EqualsQualString("issue_id") != "" {
		return listIssueWorklogs(ctx, d, h)
	}

	// Fetch worklogs by updated time. Falls back to updated > 0 in
	// cases where update filter is not provided (all worklogs)
	return listWorklogsByUpdated(ctx, d, h)
}

func listIssueWorklogs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	issueId := d.EqualsQualString("issue_id")

	if issueId == "" {
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
		apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/worklog?startAt=%d&maxResults=%d&expand=properties", issueId, last, limit)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_worklog.listIssueWorklogs", "get_request_error", err)
			return nil, err
		}

		w := new(jira.Worklog)
		_, err = client.Do(req, w)
		if err != nil {
			if isNotFoundError(err) { // Handle not found error code
				return nil, nil
			}

			plugin.Logger(ctx).Error("jira_issue_worklog.listIssueWorklogs", "api_error", err)
			return nil, err
		}

		for _, c := range w.Worklogs {
			d.StreamListItem(ctx, WorklogDetails{c, c.IssueID})

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

type ChangedWorklogs struct {
	LastPage bool            `json:"lastPage" structs:"lastPage"`
	NextPage *string         `json:"nextPage" structs:"nextPage"`
	Values   []ChangeWorklog `json:"values" structs:"values"`
}

type ChangeWorklog struct {
	WorklogId int64 `json:"WorklogId,omitempty" structs:"WorklogId,omitempty"`
}

func listWorklogsByUpdated(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	since := time.Time{}

	if d.Quals["updated"] != nil {
		for _, q := range d.Quals["updated"].Quals {
			switch q.Operator {
			case ">":
				since = q.Value.GetTimestampValue().AsTime().Add(time.Millisecond * 1)
			case ">=":
				since = q.Value.GetTimestampValue().AsTime()
			}
		}
	}

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_worklog.listWorklogsByUpdated", "connection_error", err)
		return nil, err
	}

	nextPageUrl := fmt.Sprintf("rest/api/2/worklog/updated?since=%d&expand=properties", since.UnixMilli())

	for {
		req, err := client.NewRequest("GET", nextPageUrl, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_worklog.listWorklogsByUpdated", "get_request_error", err)
			return nil, err
		}

		w := new(ChangedWorklogs)
		_, err = client.Do(req, w)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_worklog.listWorklogsByUpdated", "api_error", err)
			return nil, err
		}

		worklogIds := []int64{}
		for _, v := range w.Values {
			worklogIds = append(worklogIds, v.WorklogId)
		}

		// no need to worry about pagination in batchGetWorklog
		// since it accepts same number of worklogs per page (1000)
		// as this method returns at maximum (also 1000)
		_, err = batchGetWorklog(ctx, worklogIds, d)
		if err != nil {
			return nil, err
		}

		if w.LastPage {
			return nil, nil
		}

		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}

		if w.NextPage != nil {
			nextPageUrl = *w.NextPage
			// Extract the path from the full URL using the url package
			parsedUrl, err := url.Parse(nextPageUrl)
			if err != nil {
				plugin.Logger(ctx).Error("jira_issue_worklog.listWorklogsByUpdated", "parsing_error", err)
				return nil, err
			}
			nextPageUrl = parsedUrl.Path + "?" + parsedUrl.RawQuery
		}
	}
}

func batchGetWorklog(ctx context.Context, ids []int64, d *plugin.QueryData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_worklog.batchGetWorklog", "connection_error", err)
		return nil, err
	}

	apiEndpoint := "rest/api/2/worklog/list?expand=properties"

	body := struct {
		Ids []int64 `json:"ids"`
	}{
		Ids: ids,
	}

	req, err := client.NewRequest("POST", apiEndpoint, body)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_worklog.batchGetWorklog", "get_request_error", err)
		return nil, err
	}

	w := make([]jira.WorklogRecord, 0)
	_, err = client.Do(req, &w)

	if err != nil {
		if isNotFoundError(err) { // Handle not found error code
			return nil, nil
		}

		plugin.Logger(ctx).Error("jira_issue_worklog.batchGetWorklog", "api_error", err)
		return nil, err
	}

	for _, c := range w {
		d.StreamListItem(ctx, WorklogDetails{c, c.IssueID})

		// Context may get cancelled due to manual cancellation or if the limit has been reached
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
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
