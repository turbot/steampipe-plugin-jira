package jira

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/trivago/tgo/tcontainer"
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
				{Name: "component", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "labels", Require: plugin.Optional, Operators: []string{"=", "<>", "~~"}},
				{Name: "duedate", Require: plugin.Optional, Operators: []string{"=", ">", ">=", "<=", "<"}},
				{Name: "epic_key", Require: plugin.Optional, Operators: []string{"=", "<>"}},
				{Name: "parent_key", Require: plugin.Optional, Operators: []string{"=", "<>"}},
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
				Transform:   transform.FromGo().Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "key",
				Description: "The key of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo().Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "self",
				Description: "The URL of the issue details.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo().Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "project_key",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Key").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "project_id",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.ID").Transform(lowerIfCaseInsensitive),
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
				Transform:   transform.FromField("Fields.Status.StatusCategory.Name").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "epic_key",
				Description: "The key of the epic to which issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromP(extractRequiredField, "epic").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "parent_key",
				Description: "The key of the epic to which issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Parent.Key").Transform(lowerIfCaseInsensitive),
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
				Transform:   transform.FromField("Fields.Assignee.AccountID").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "assignee_display_name",
				Description: "Display name the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Assignee.DisplayName").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "creator_account_id",
				Description: "Account Id of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.AccountID").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "creator_display_name",
				Description: "Display name of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.DisplayName").Transform(lowerIfCaseInsensitive),
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
				Transform:   transform.FromField("Fields.Description").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "type",
				Description: "The name of the issue type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Type.Name").Transform(lowerIfCaseInsensitive),
			},
			/*{
				Name:        "labels",
				Description: "A list of labels applied to the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Labels"),
			},*/
			{
				Name:        "priority",
				Description: "Priority assigned to the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Priority.Name").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "project_name",
				Description: "Name of the project to that issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Name").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "reporter_account_id",
				Description: "Account Id of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.AccountID").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "reporter_display_name",
				Description: "Display name of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.DisplayName").Transform(lowerIfCaseInsensitive),
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
				Transform:   transform.FromField("Fields.Summary").Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "updated",
				Description: "Time when the issue was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Updated").Transform(convertJiraTime),
			},

			// JSON fields
			{
				Name:        "labels",
				Description: "A list of labels applied to the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(getIssueLabels).Transform(lowerIfCaseInsensitive),
				//Transform:   transform.From("Fields.Components").Transform(extractLabels).Transform(convertToCsv).Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "component",
				Description: "List of components Name associated with the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Components").Transform(extractComponentNames).Transform(lowerIfCaseInsensitive),
				//Transform:   transform.From("Fields.Components").Transform(extractComponentNames).Transform(convertToCsv).Transform(lowerIfCaseInsensitive),
			},
			{
				Name:        "components",
				Description: "List of components associated with the issue.",
				Type:        proto.ColumnType_STRING,
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
				Transform:   transform.FromField("Key").Transform(lowerIfCaseInsensitive),
			},
		},
	}
}

//// LIST FUNCTION

func listIssues(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	useExpression := true
	columnRequiresJQL := map[string]struct{}{}
	columnRequiresJQL["sprint_ids"] = struct{}{}
	columnRequiresJQL["sprint_names"] = struct{}{}
	columnRequiresJQL["epic_key"] = struct{}{}
	columnRequiresJQL["tags"] = struct{}{}
	columnRequiresJQL["components"] = struct{}{}
	for _, column := range d.QueryContext.Columns {
		if _, ok := columnRequiresJQL[column]; ok {
			useExpression = false
			break
		}
	}
	plugin.Logger(ctx).Debug("jira_issue.listIssues", "useExpression", useExpression)

	last := 0
	issueCount := 1
	numLoops := 5
	issueLimit, err := getIssueLimit(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.listIssues", "issue_limit", err)
		return nil, err
	}
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
	jql := buildJQLQueryFromQuals(ctx, d.Quals, d.Table.Columns)
	// set options.MaxResults to the smaller of user-defined limit and calculated
	// maxResults value
	if useExpression {
		if maxResults, err := calculateMaxResults(ctx, d, jql); err != nil {
			return nil, err
		} else if queryLimit != nil && int(*queryLimit) < maxResults {
			options.MaxResults = int(*queryLimit)
		} else {
			options.MaxResults = maxResults
		}
	}

	for {
		var searchResult *searchResult
		var res *jira.Response
		var err error
		if useExpression {
			searchResult, res, err = searchWithExpression(ctx, d, jql, &options)
			if searchResult.MaxResults < options.MaxResults {
				plugin.Logger(ctx).Debug("jira_issue.listIssues", "maxResults < options.MaxResults; lowering", searchResult.MaxResults)
				options.MaxResults = searchResult.MaxResults
			}
			issueLimit = searchResult.MaxResults * numLoops
		} else {
			searchResult, res, err = searchWithContext(ctx, d, jql, &options)
		}
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue.listIssues", "search_error", err)
			return nil, err
		}

		issues := searchResult.Issues
		var names map[string]string
		if !useExpression {
			names = searchResult.Names
		}
		body, _ := io.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_issue.listIssues", "res_body", string(body))

		if err != nil {
			if isNotFoundError(err) || isBadRequestError(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("jira_issue.listIssues", "api_error", err)
			return nil, err
		}

		// return error if user requests too much data
		if searchResult.Total > issueLimit {
			//return nil, errors.New(fmt.Sprintf("Number of results exceeds issue limit(%d>%d). Please make your query more specific.", searchResult.Total, issueLimit))
			plugin.Logger(ctx).Debug(fmt.Sprintf("Number of results exceeds issue limit(%d>%d). Please make your query more specific.", searchResult.Total, issueLimit))
		}

		sensitivity, err := getCaseSensitivity(ctx, d)
		if err != nil {
			return nil, err
		}
		plugin.Logger(ctx).Debug("jira_issue.listIssues", "case_sensitivity", sensitivity)

		for _, issue := range issues {
			if issueCount > issueLimit {
				plugin.Logger(ctx).Debug("Maximum number of issues reached", issueLimit)
				return nil, nil
			}

			issue.Fields.Unknowns["sensitivity"] = sensitivity

			// plugin.Logger(ctx).Debug("Issue output:", issue)
			// plugin.Logger(ctx).Debug("Issue names output:", names)
			var keys map[string]string
			if !useExpression {
				keys = map[string]string{
					"epic":   getFieldKey(ctx, d, names, "Epic Link"),
					"sprint": getFieldKey(ctx, d, names, "Sprint"),
				}
			}
			d.StreamListItem(ctx, IssueInfo{issue, keys})
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
			issueCount += 1
		}

		last = searchResult.StartAt + len(issues)
		if last >= searchResult.Total {
			return nil, nil
		} else if issueCount >= issueLimit {
			plugin.Logger(ctx).Debug("Maximum number of issues reached", issueLimit)
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

	if sensitivity, err := getCaseSensitivity(ctx, d); err != nil {
		return nil, err
	} else {
		plugin.Logger(ctx).Debug("jira_issue.getIssue", "case_sensitivity", sensitivity)
		issue.Fields.Unknowns["sensitivity"] = sensitivity
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
	status := d.EqualsQualString("status")
	sensitivity := issue.Fields.Unknowns["sensitivity"]

	if status == "" {
		status = issue.Fields.Status.Name
	}

	if sensitivity == "sensitive" {
		return status, nil
	} else {
		return strings.ToLower(status), nil
	}
}

//// TRANSFORM FUNCTION

func extractComponentIds(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	var componentIds []string
	for _, item := range d.Value.([]*jira.Component) {
		//plugin.Logger(ctx).Debug("extractComponentIds", item)
		componentIds = append(componentIds, item.ID)
	}
	return componentIds, nil
}

func extractComponentNames(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	var componentNames []string
	for _, item := range d.Value.([]*jira.Component) {
		componentNames = append(componentNames, item.Name)
	}
	return strings.Join(componentNames, ","), nil
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

func getIssueLabels(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issue := d.HydrateItem.(IssueInfo)
	return strings.Join(issue.Fields.Labels, ","), nil
}

func getIssueComponents(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	issue := d.HydrateItem.(IssueInfo)

	var componentNames []string
	if issue.Fields != nil && issue.Fields.Components != nil {
		for _, i := range issue.Fields.Components {
			plugin.Logger(ctx).Debug("getIssueComponents", i)
			componentNames = append(componentNames, i.Name)
		}
	}
	result := strings.Join(componentNames, ",")
	return result, nil

}

//// Utility Function

// getFieldKey:: get key for unknown expanded fields
func getFieldKey(ctx context.Context, d *plugin.QueryData, names map[string]string, keyName string) string {

	// plugin.Logger(ctx).Debug("Check for keyName", names)
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

// lowerIfCaseInsensitive:: used for columns of type proto.ColumnType_STRING
// attempts to convert the value to lowercase if it is not nil, otherwise returns nil
func lowerIfCaseInsensitive(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	issue := d.HydrateItem.(IssueInfo)

	sensitivity := issue.Fields.Unknowns["sensitivity"]
	if sensitivity == "sensitive" {
		return d.Value, nil
	}

	if str, ok := d.Value.(string); ok {
		return strings.ToLower(str), nil
	} else if d.Value == nil {
		return d.Value, nil
	}

	return d.Value, errors.New("Could not transform field value to lowercase.")
}

func searchWithExpression(ctx context.Context, d *plugin.QueryData, jql string, options *jira.SearchOptions) (*searchResult, *jira.Response, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.listIssues.searchWithExpression", "connection_error", err)
		return nil, nil, err
	}

	jiraConfig := GetConfig(d.Connection)

	u := url.URL{
		Path: "rest/api/3/expression/eval",
	}
	uv := url.Values{}
	uv.Add("expand", "meta.complexity")
	u.RawQuery = uv.Encode()

	if jql == "" {
		jql = "order by created DESC"
	}
	requestBody := map[string]interface{}{
		"context": map[string]interface{}{
			"issues": map[string]interface{}{
				"jql": map[string]interface{}{
					"query":      jql,
					"maxResults": options.MaxResults,
					"startAt":    options.StartAt,
				},
			},
		},
		"expression": "issues.map(issue => {" + getKeyString(ctx, d.QueryContext.Columns) + "})",
	}

	plugin.Logger(ctx).Debug("jira_issue.listIssues.searchWithExpression", "req_body", requestBody)
	req, err := client.NewRequestWithContext(ctx, "POST", u.String(), requestBody)
	if err != nil {
		return nil, nil, err
	}

	expressionResult := new(issueExpressionResult)
	resp, err := client.Do(req, expressionResult)
	body, _ := io.ReadAll(resp.Body)
	plugin.Logger(ctx).Debug("jira_issue.listIssues.searchWithExpression", "res_body", string(body))
	if err != nil {
		return nil, nil, jira.NewJiraError(resp, err)
	}

	// convert expressionResults to jira issues
	var jiraIssues []jira.Issue
	for _, value := range expressionResult.Values {
		n := new(jira.Issue)
		f := new(jira.IssueFields)
		var components []*jira.Component
		for _, component := range value.Components {
			components = append(components, &jira.Component{ID: component["id"], Name: component["name"]})
		}
		f.Components = components
		f.Parent = &jira.Parent{Key: value.ParentKey}

		timeLayout := "2006-01-02T15:04:05.999-0700"
		created, _ := time.Parse(timeLayout, value.Created)
		f.Created = jira.Time(created)

		resolution, _ := time.Parse(timeLayout, value.ResolutionDate)
		f.Resolutiondate = jira.Time(resolution)

		updated, _ := time.Parse(timeLayout, value.Updated)
		f.Updated = jira.Time(updated)

		duedate, _ := time.Parse(timeLayout, value.Duedate)
		f.Duedate = jira.Date(duedate)

		f.Assignee = &jira.User{DisplayName: value.AssigneeName, AccountID: value.AssigneeID}
		f.Creator = &jira.User{DisplayName: value.CreatorName, AccountID: value.CreatorID}
		f.Description = value.Description
		f.Labels = value.Labels
		f.Priority = &jira.Priority{Name: value.Priority}
		f.Project = jira.Project{ID: value.ProjectID, Key: value.ProjectKey, Name: value.ProjectName}
		f.Reporter = &jira.User{DisplayName: value.ReporterName, AccountID: value.ReporterID}
		f.Status = &jira.Status{Name: value.StatusName, StatusCategory: jira.StatusCategory{Name: value.StatusCategory}}
		f.Summary = value.Summary
		f.Type = jira.IssueType{Name: value.Type}
		f.Unknowns = make(tcontainer.MarshalMap)

		n.ID = value.ID
		n.Key = value.Key
		n.Self = strings.TrimSuffix(*jiraConfig.BaseUrl, "/") + "/rest/api/2/issue/" + n.ID
		n.Fields = f
		jiraIssues = append(jiraIssues, *n)
	}

	v := new(searchResult)
	v.StartAt = expressionResult.Meta.Issues.Jql.StartAt
	v.MaxResults = expressionResult.Meta.Issues.Jql.MaxResults
	v.Total = expressionResult.Meta.Issues.Jql.TotalCount
	v.Issues = jiraIssues

	return v, resp, err
}

// generate expression key string from columns in d.QueryContext.Columns
func getKeyString(ctx context.Context, columns []string) string {
	columnMapping := map[string]string{
		"id":                    "id: JSON.stringify(issue.id)",
		"key":                   "key: issue.key",
		"project_name":          "projectName: issue.project.name",
		"project_id":            "projectId: JSON.stringify(issue.project.id)",
		"project_key":           "projectKey: issue.project.key",
		"status":                "statusName: issue.status.name",
		"status_category":       "statusCategory: issue.status.category.name",
		"assignee_account_id":   "assigneeId: JSON.stringify(issue.assignee?.accountId)",
		"assignee_display_name": "assigneeName: issue.assignee?.displayName",
		"creator_account_id":    "creatorId: JSON.stringify(issue.creator?.accountId)",
		"creator_display_name":  "creatorName: issue.creator?.displayName",
		"created":               "created: issue.created",
		"duedate":               "dueDate: issue.dueDate",
		"description":           "description: issue.description?.plainText",
		"type":                  "issueType: issue.issueType.name",
		"labels":                "labels: issue.labels",
		"priority":              "priority: issue.priority.name",
		"reporter_display_name": "reporterName: issue.reporter?.displayName",
		"reporter_account_id":   "reporterId: JSON.stringify(issue.reporter?.accountId)",
		"resolution_date":       "resolutionDate: issue.resolutionDate",
		"summary":               "summary: issue.summary",
		"updated":               "updated: issue.updated",
		"parent_key":            "parentKey: issue.parent?.key",
		"component":             "components: issue.components.map(c => { id: JSON.stringify(c.id), name: c.name }) ",
		//"components": "components: issue.components?.map(c => { id: JSON.stringify(c.id), name: c.name }) ",
	}
	keys := []string{}
	for _, column := range columns {
		if key, ok := columnMapping[column]; ok {
			keys = append(keys, key)
		} else {
			plugin.Logger(ctx).Debug("jira_issue.listIssues.searchWithExpression.getKeyString", "column not found in mapping", column)
		}
	}
	return strings.Join(keys, ",")
}

func calculateMaxResults(ctx context.Context, d *plugin.QueryData, jql string) (int, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.listIssues.searchWithExpression", "connection_error", err)
		return 0, err
	}

	u := url.URL{
		Path: "rest/api/3/expression/eval",
	}
	uv := url.Values{}
	uv.Add("expand", "meta.complexity")
	u.RawQuery = uv.Encode()
	url := u.String()

	if jql == "" {
		jql = "order by created DESC"
	}
	resultAmount := 2
	requestBody := map[string]interface{}{
		"context": map[string]interface{}{
			"issues": map[string]interface{}{
				"jql": map[string]interface{}{
					"query":      jql,
					"maxResults": resultAmount,
				},
			},
		},
		"expression": "issues.map(issue => {" + getKeyString(ctx, d.QueryContext.Columns) + "})",
	}

	plugin.Logger(ctx).Debug("jira_issue.listIssues.searchWithExpression.calculateMaxResults", "req_body", requestBody)
	req, err := client.NewRequestWithContext(ctx, "POST", url, requestBody)
	if err != nil {
		return 0, err
	}

	expressionResult := new(issueExpressionResult)
	resp, err := client.Do(req, expressionResult)
	body, _ := io.ReadAll(resp.Body)
	plugin.Logger(ctx).Debug("jira_issue.listIssues.searchWithExpression.calculateMaxResults", "res_body", string(body))
	if err != nil {
		return 0, jira.NewJiraError(resp, err)
	}

	plugin.Logger(ctx).Debug("jira_issue.listIssues.searchWithExpression.calculateMaxResults", "complexity", expressionResult.Meta.Complexity)
	primitiveValuePortion := float64(expressionResult.Meta.Complexity.PrimitiveValues.Value) / float64(resultAmount)
	primitiveValueMax := float64(expressionResult.Meta.Complexity.PrimitiveValues.Limit)/primitiveValuePortion - primitiveValuePortion

	stepPortion := float64(expressionResult.Meta.Complexity.Steps.Value) / float64(resultAmount)
	stepMax := float64(expressionResult.Meta.Complexity.Steps.Limit)/stepPortion - stepPortion

	if primitiveValueMax < stepMax {
		return int(primitiveValueMax), nil
	} else {
		return int(stepMax), nil
	}
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

	// Specify fields to prevent getting more data than necessary
	fields := []string{
		"project",
		"status",
		"assignee",
		"creator",
		"created",
		"duedate",
		"description",
		"issuetype",
		"labels",
		"priority",
		"reporter",
		"resolutiondate",
		"summary",
		"updated",
		"components",
	}
	fieldString := strings.Join(fields, ",")
	uv.Add("fields", fieldString)

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

type issueExpressionValue struct {
	ID             string              `json:"id,omitempty" structs:"id,omitempty"`
	Key            string              `json:"key,omitempty" structs:"key,omitempty"`
	Self           string              `json:"self,omitempty" structs:"self,omitempty"`
	Summary        string              `json:"summary,omitempty" structs:"summary,omitempty"`
	Type           string              `json:"issueType,omitempty" structs:"issueType,omitempty"`
	CreatorID      string              `json:"creatorId,omitempty" structs:"creatorId,omitempty"`
	CreatorName    string              `json:"creatorName,omitempty" structs:"creatorName,omitempty"`
	Components     []map[string]string `json:"components,omitempty" structs:"components,omitempty"`
	Created        string              `json:"created,omitempty" structs:"created,omitempty"`
	ProjectName    string              `json:"projectName,omitempty" structs:"projectName,omitempty"`
	ProjectID      string              `json:"projectId,omitempty" structs:"projectId,omitempty"`
	ProjectKey     string              `json:"projectKey,omitempty" structs:"projectKey,omitempty"`
	Description    string              `json:"description,omitempty" structs:"description,omitempty"`
	ReporterName   string              `json:"reporterName,omitempty" structs:"reporterName,omitempty"`
	ReporterID     string              `json:"reporterId,omitempty" structs:"reporterId,omitempty"`
	Priority       string              `json:"priority,omitempty" structs:"priority,omitempty"`
	Labels         []string            `json:"labels,omitempty" structs:"labels,omitempty"`
	Duedate        string              `json:"dueDate,omitempty" structs:"dueDate,omitempty"`
	ResolutionDate string              `json:"resolutionDate,omitempty" structs:"resolutionDate,omitempty"`
	AssigneeID     string              `json:"assigneeId,omitempty" structs:"assigneeId,omitempty"`
	AssigneeName   string              `json:"assigneeName,omitempty" structs:"assigneeName,omitempty"`
	Updated        string              `json:"updated,omitempty" structs:"updated,omitempty"`
	StatusName     string              `json:"statusName,omitempty" structs:"statusName,omitempty"`
	StatusCategory string              `json:"statusCategory,omitempty" structs:"statusCategory,omitempty"`
	ParentKey      string              `json:"parentKey,omitempty" structs:"parentKey,omitempty"`
}

type issueExpressionResult struct {
	Values []issueExpressionValue `json:"value" structs:"value"`
	Meta   struct {
		Complexity struct {
			Steps struct {
				Value int `json:"value" structs:"value"`
				Limit int `json:"limit" structs:"limit"`
			} `json:"steps" structs:"steps"`
			ExpensiveOperations struct {
				Value int `json:"value" structs:"value"`
				Limit int `json:"limit" structs:"limit"`
			} `json:"expensiveOperations" structs:"expensiveOperations"`
			Beans struct {
				Value int `json:"value" structs:"value"`
				Limit int `json:"limit" structs:"limit"`
			} `json:"beans" structs:"beans"`
			PrimitiveValues struct {
				Value int `json:"value" structs:"value"`
				Limit int `json:"limit" structs:"limit"`
			} `json:"primitiveValues" structs:"primitiveValues"`
		} `json:"complexity" structs:"complexity"`
		Issues struct {
			Jql struct {
				StartAt    int `json:"startAt" structs:"startAt"`
				MaxResults int `json:"maxResults" structs:"maxResults"`
				Count      int `json:"count" structs:"count"`
				TotalCount int `json:"totalCount" structs:"totalCount"`
			} `json:"jql" structs:"jql"`
		} `json:"issues" structs:"issues"`
	} `json:"meta" structs:"meta"`
}
