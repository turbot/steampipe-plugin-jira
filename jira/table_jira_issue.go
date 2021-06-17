package jira

import (
	"context"
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableIssue(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_issue",
		Description: "Issues help manage code, estimate workload, and keep track of team.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getIssue,
		},
		List: &plugin.ListConfig{
			Hydrate: listIssues,
		},
		Columns: []*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "key",
				Description: "The key of the issue.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the issue details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_key",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Key"),
			},
			{
				Name:        "project_id",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.ID"),
			},
			{
				Name:        "status",
				Description: "Json object containing important subfields info the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Status.Name"),
			},
			{
				Name:        "epic_key",
				Description: "The key of the epic to which issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromP(extractRequiredField, "epic"),
			},
			{
				Name:        "sprint_ids",
				Description: "The list of ids of the sprint to which issue belongs.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromP(extractRequiredField, "sprint").Transform(extractSprintIds),
			},
			{
				Name:        "sprint_names",
				Description: "The list of names of the sprint to which issue belongs.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromP(extractRequiredField, "sprint").Transform(extractSprintNames),
			},

			// other important fields
			{
				Name:        "assignee_account_id",
				Description: "Account Id the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Assignee.AccountID"),
			},
			{
				Name:        "assignee_display_name",
				Description: "Display name the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Assignee.DisplayName"),
			},
			{
				Name:        "creator_account_id",
				Description: "Account Id of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.AccountID"),
			},
			{
				Name:        "creator_display_name",
				Description: "Display name of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.DisplayName"),
			},
			{
				Name:        "created",
				Description: "Time when the issue was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Created").Transform(convertJiraTime),
			},
			{
				Name:        "duedate",
				Description: "Time by which the issue is expected to be completed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Duedate").NullIfZero().Transform(convertJiraDate),
			},
			{
				Name:        "description",
				Description: "Description of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Description"),
			},
			{
				Name:        "type",
				Description: "The name of the issue type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Type.Name"),
			},
			{
				Name:        "labels",
				Description: "A list of labels applied to the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Labels"),
			},
			{
				Name:        "priority",
				Description: "Priority assigned to the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Priority.Name"),
			},
			{
				Name:        "project_name",
				Description: "Name of the project to that issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Name"),
			},
			{
				Name:        "reporter_account_id",
				Description: "Account Id of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.AccountID"),
			},
			{
				Name:        "reporter_display_name",
				Description: "Display name of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.DisplayName"),
			},
			{
				Name:        "summary",
				Description: "Details of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Summary"),
			},
			{
				Name:        "updated",
				Description: "Time when the issue was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Updated").Transform(convertJiraTime),
			},

			// JSON fields
			{
				Name:        "components",
				Description: "List of components associated with the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Components").Transform(extractComponentIds),
			},
			{
				Name:        "fields",
				Description: "Json object containing important subfields of the issue.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "tags",
				Type:        proto.ColumnType_JSON,
				Description: "A map of label names associated with this issue, in Steampipe standard format.",
				Transform:   transform.From(getIssueTags),
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Key"),
			},
		},
	}
}

//// LIST FUNCTION

func listIssues(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("listIssues")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	var epicKey, sprintKey string
	for {

		apiEndpoint := fmt.Sprintf(
			"/rest/api/2/search?startAt=%d&maxResults=%d&expand=names",
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			if isNotFoundError(err) || strings.Contains(err.Error(), "400") {
				return nil, nil
			}
			logger.Error("listIssues", "Error", err)
			return nil, err
		}

		listIssuesResult := new(ListIssuesResult)
		_, err = client.Do(req, listIssuesResult)
		if err != nil {
			return nil, err
		}

		epicKey = getFieldKey(ctx, d, listIssuesResult.Names, "Epic Link")
		sprintKey = getFieldKey(ctx, d, listIssuesResult.Names, "Sprint")

		keys := map[string]string{
			"epic":   epicKey,
			"sprint": sprintKey,
		}

		for _, issue := range listIssuesResult.Issues {
			d.StreamListItem(ctx, IssueInfo{issue, keys})
		}

		last = listIssuesResult.StartAt + len(listIssuesResult.Issues)
		if last >= listIssuesResult.Total {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getIssue(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getIssue")

	issueId := d.KeyColumnQuals["id"].GetStringValue()

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	issue, _, err := client.Issue.Get(issueId, &jira.GetQueryOptions{
		Expand: "names",
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		logger.Error("getIssue", "Error", err)
		return nil, err
	}

	epicKey := getFieldKey(ctx, d, issue.Names, "Epic Link")
	sprintKey := getFieldKey(ctx, d, issue.Names, "Sprint")

	keys := map[string]string{
		"epic":   epicKey,
		"sprint": sprintKey,
	}

	return IssueInfo{*issue, keys}, err
}

//// TRANSFORM FUNCTION

func extractComponentIds(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var componentIds []string
	for _, item := range d.Value.([]*jira.Component) {
		componentIds = append(componentIds, item.ID)
	}
	return componentIds, nil
}

func extractRequiredField(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issueInfo := d.HydrateItem.(IssueInfo)
	m := issueInfo.Fields.Unknowns
	param := d.Param.(string)
	return m[issueInfo.Keys[param]], nil
}

func extractSprintIds(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	var sprintIds []interface{}
	for _, item := range d.Value.([]interface{}) {
		sprint := item.(map[string]interface{})
		sprintIds = append(sprintIds, sprint["id"])
	}

	return sprintIds, nil
}
func extractSprintNames(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	var sprintNames []interface{}
	for _, item := range d.Value.([]interface{}) {
		sprint := item.(map[string]interface{})
		sprintNames = append(sprintNames, sprint["name"])
	}

	return sprintNames, nil
}

func getIssueTags(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issue := d.HydrateItem.(IssueInfo)

	tags := make(map[string]bool)
	if issue.Fields != nil && issue.Fields.Labels != nil {
		for _, i := range issue.Fields.Labels {
			tags[i] = true
		}
	}
	return tags, nil
}

//// Utility Function

// getFieldKey:: get key for unknown expanded fields
func getFieldKey(ctx context.Context, d *plugin.QueryData, names map[string]string, keyName string) string {
	cacheKey := "issue-" + keyName
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.(string)
	}

	for k, v := range names {
		if v == keyName {
			d.ConnectionManager.Cache.Set(cacheKey, k)
			return k
		}
	}
	return ""
}

//// Required Structs

type ListIssuesResult struct {
	Expand     string            `json:"expand"`
	MaxResults int               `json:"maxResults"`
	StartAt    int               `json:"startAt"`
	Total      int               `json:"total"`
	Issues     []jira.Issue      `json:"issues"`
	Names      map[string]string `json:"names,omitempty" structs:"names,omitempty"`
}

type IssueInfo struct {
	jira.Issue
	Keys map[string]string
}
