package jira

import (
	"context"
	"io"
	"net/url"
	"strconv"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableIssue(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_issue",
		Description: "Issues help manage code, estimate workload, and keep track of team.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AnyColumn([]string{"id", "key"}),
			Hydrate:    getIssue,
		},
		List: &plugin.ListConfig{
			Hydrate: listIssues,
			// https://support.atlassian.com/jira-service-management-cloud/docs/advanced-search-reference-jql-fields/
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "assignee_account_id", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "assignee_display_name", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "created", Require: plugin.Optional, Operators: []string{"=", ">", ">=", "<=", "<"}},
				{Name: "creator_account_id", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "creator_display_name", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "duedate", Require: plugin.Optional, Operators: []string{"=", ">", ">=", "<=", "<"}},
				{Name: "epic_key", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "priority", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "project_id", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "project_key", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "project_name", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "reporter_account_id", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "reporter_display_name", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "resolution_date", Require: plugin.Optional, Operators: []string{"=", ">", ">=", "<=", "<"}},
				{Name: "status", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "status_category", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "type", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "updated", Require: plugin.Optional, Operators: []string{"=", ">", ">=", "<=", "<"}},
			},
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
				Hydrate:     getStatusValue,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "status_category",
				Description: "The status category (Open, In Progress, Done) of the ticket.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Status.StatusCategory.Name"),
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
				Name:        "resolution_date",
				Description: "Date the issue was resolved.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Resolutiondate").Transform(convertJiraTime),
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

	last := 0
	pageSize, err := getPageSize(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.listIssues", "page_size", err)
		return nil, err
	}
	plugin.Logger(ctx).Debug("jira_issue.listIssues", "page_size", pageSize)

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var limit int = pageSize
	if d.QueryContext.Limit != nil {
		if *queryLimit < int64(limit) {
			limit = int(*queryLimit)
		}
	}
	options := jira.SearchOptions{
		StartAt:    0,
		MaxResults: limit,
		Expand:     "names",
	}

	jql := buildJQLQueryFromQuals(d.Quals, d.Table.Columns)
	plugin.Logger(ctx).Debug("jira_issue.listIssues", "JQL", jql)

	for {
		searchResult, res, err := searchWithContext(ctx, d, jql, &options)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue.listIssues", "search_error", err)
			return nil, err
		}
		issues := searchResult.Issues
		names := searchResult.Names
		body, _ := io.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_issue.listIssues", "res_body", string(body))

		if err != nil {
			if isNotFoundError(err) || isBadRequestError(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("jira_issue.listIssues", "api_error", err)
			return nil, err
		}

		for _, issue := range issues {
			plugin.Logger(ctx).Debug("Issue output:", issue)
			plugin.Logger(ctx).Debug("Issue names output:", names)
			keys := map[string]string{
				"epic":   getFieldKey(ctx, d, names, "Epic Link"),
				"sprint": getFieldKey(ctx, d, names, "Sprint"),
			}
			d.StreamListItem(ctx, IssueInfo{issue, keys})
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		last = searchResult.StartAt + len(issues)
		if last >= searchResult.Total {
			return nil, nil
		} else {
			options.StartAt = last
		}
	}
}

//// HYDRATE FUNCTION

func getIssue(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Debug("getIssue")

	issueId := d.EqualsQualString("id")
	key := d.EqualsQualString("key")

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.getIssue", "connection_error", err)
		return nil, err
	}

	var id string
	if issueId != "" {
		id = issueId
	} else if key != "" {
		id = key
	} else {
		return nil, nil
	}

	issue, res, err := client.Issue.Get(id, &jira.GetQueryOptions{
		Expand: "names",
	})
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_issue.getIssue", "res_body", string(body))
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_issue.getIssue", "api_error", err)
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

func getStatusValue(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	issue := h.Item.(IssueInfo)
	issueStauus := d.EqualsQualString("status")

	if issueStauus != "" {
		return issueStauus, nil
	}
	return issue.Fields.Status.Name, nil
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
		if sprint, ok := item.(map[string]interface{}); ok {
			sprintIds = append(sprintIds, sprint["id"])
		}
	}

	return sprintIds, nil
}
func extractSprintNames(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	var sprintNames []interface{}
	for _, item := range d.Value.([]interface{}) {
		if sprint, ok := item.(map[string]interface{}); ok {
			sprintNames = append(sprintNames, sprint["name"])
		}
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

	plugin.Logger(ctx).Debug("Check for keyName", names)
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

func searchWithContext(ctx context.Context, d *plugin.QueryData, jql string, options *jira.SearchOptions) (*searchResult, *jira.Response, error) {
	u := url.URL{
		Path: "rest/api/2/search",
	}
	uv := url.Values{}
	if jql != "" {
		uv.Add("jql", jql)
	}

	// Append the values of options to the path parameters
	if options.StartAt != 0 {
		uv.Add("startAt", strconv.Itoa(options.StartAt))
	}
	if options.MaxResults != 0 {
		uv.Add("maxResults", strconv.Itoa(options.MaxResults))
	}
	uv.Add("expand", options.Expand)

	u.RawQuery = uv.Encode()

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.listIssues.searchWithContext", "connection_error", err)
		return nil, nil, err
	}

	req, err := client.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	v := new(searchResult)
	resp, err := client.Do(req, v)
	body, _ := io.ReadAll(resp.Body)
	plugin.Logger(ctx).Debug("jira_issue.listIssues.searchWithContext", "res_body", string(body))
	if err != nil {
		err = jira.NewJiraError(resp, err)
	}
	return v, resp, err
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

type searchResult struct {
	Issues     []jira.Issue      `json:"issues" structs:"issues"`
	StartAt    int               `json:"startAt" structs:"startAt"`
	MaxResults int               `json:"maxResults" structs:"maxResults"`
	Total      int               `json:"total" structs:"total"`
	Names      map[string]string `json:"names,omitempty" structs:"names,omitempty"`
}
