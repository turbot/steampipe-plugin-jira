package jira

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
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
		Columns: []*plugin.Column{
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
		},
	}
}

//// LIST FUNCTION

func listIssueTypes(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("listIssueTypes")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	req, err := client.NewRequest("GET", "/rest/api/3/issuetype", nil)
	if err != nil {
		if isNotFoundError(err) || strings.Contains(err.Error(), "400") {
			return nil, nil
		}
		logger.Error("listIssueTypes", "Error", err)
		return nil, err
	}

	issuesTypeResult := new([]ListIssuesTypeResult)
	_, err = client.Do(req, issuesTypeResult)
	if err != nil {
		return nil, err
	}

	for _, issueType := range *issuesTypeResult {
		d.StreamListItem(ctx, issueType)
	}

	return nil, err
}

//// HYDRATE FUNCTION

func getIssueType(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getIssueType")

	issueTypeID := d.KeyColumnQuals["id"].GetStringValue()

	if issueTypeID == "" {
		return nil, nil
	}
	issueType := new(ListIssuesTypeResult)
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/issuetype/%s", issueTypeID)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	_, err = client.Do(req, issueType)
	if err != nil && isNotFoundError(err) {
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
