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

func tableIssueType(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_issue_type",
		Description: "Issue types distinguish different types of work in unique ways, and help you identify, categorize, and report on your team’s work across your Jira site.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getIssueType,
		},
		List: &plugin.ListConfig{
			Hydrate: listIssueTypes,
		},
		Columns: commonColumns([]*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the issue type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The name of the issue type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the issue type details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the issue type.",
				Type:        proto.ColumnType_STRING,
			},

			// other important fields
			{
				Name:        "avatar_id",
				Description: "The ID of the issue type's avatar.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "entity_id",
				Description: "Unique ID for next-gen projects.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "hierarchy_level",
				Description: "Hierarchy level of the issue type.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "icon_url",
				Description: "The URL of the issue type's avatar.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "subtask",
				Description: "Whether this issue type is used to create subtasks.",
				Type:        proto.ColumnType_BOOL,
			},

			// JSON fields
			{
				Name:        "scope",
				Description: "Details of the next-gen projects the issue type is available in.",
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

func listIssueTypes(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_type.listIssueTypes", "connection_error", err)
		return nil, err
	}

	// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-types/
	// Paging not supported
	req, err := client.NewRequest("GET", "/rest/api/3/issuetype", nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_type.listIssueTypes", "get_request_error", err)
		return nil, err
	}

	issuesTypeResult := new([]ListIssuesTypeResult)
	res, err := client.Do(req, issuesTypeResult)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_issue_type.listIssueTypes", "res_body", string(body))

	if err != nil {
		if isNotFoundError(err) || isBadRequestError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_issue_type.listIssueTypes", "api_error", err)
		return nil, err
	}

	for _, issueType := range *issuesTypeResult {
		d.StreamListItem(ctx, issueType)
		// Context may get cancelled due to manual cancellation or if the limit has been reached
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, err
}

//// HYDRATE FUNCTION

func getIssueType(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	issueTypeID := d.EqualsQuals["id"].GetStringValue()

	if issueTypeID == "" {
		return nil, nil
	}
	issueType := new(ListIssuesTypeResult)
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_type.getIssueType", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/issuetype/%s", issueTypeID)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_type.getIssueType", "get_request_error", err)
		return nil, err
	}

	res, err := client.Do(req, issueType)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_issue_type.getIssueType", "res_body", string(body))

	if err != nil && isNotFoundError(err) {
		plugin.Logger(ctx).Error("jira_issue_type.getIssueType", "api_error", err)
		return nil, nil
	}

	return issueType, err
}

//// Required Structs

type ListIssuesTypeResult struct {
	Self           string         `json:"self"`
	ID             string         `json:"id"`
	Description    string         `json:"description"`
	IconURL        string         `json:"iconUrl"`
	Name           string         `json:"name"`
	Subtask        bool           `json:"subtask"`
	AvatarID       int64          `json:"avatarId"`
	EntityID       int64          `json:"entityId"`
	HierarchyLevel int32          `json:"hierarchyLevel"`
	Scope          IssueTypeScope `json:"scope"`
}

type IssueTypeScope struct {
	Type    string           `json:"type"`
	Project IssueTypeProject `json:"project"`
}

type IssueTypeProject struct {
	ID string `json:"id"`
}
