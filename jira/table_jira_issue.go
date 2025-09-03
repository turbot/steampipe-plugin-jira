package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
			ParentHydrate: listProjects,
			Hydrate:       listIssues,
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
		Columns: commonColumns([]*plugin.Column{
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
				Transform:   transform.FromField("V3Issue.Fields.Project.Key"),
			},
			{
				Name:        "project_id",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Project.ID"),
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
				Transform:   transform.FromField("V3Issue.Fields.StatusCategory.Name"),
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
				Transform:   transform.FromField("V3Issue.Fields.Sprints").Transform(extractSprintIds),
			},
			{
				Name:        "sprint_names",
				Description: "The list of names of the sprint to which issue belongs.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("V3Issue.Fields.Sprints").Transform(extractSprintNames),
			},

			// other important fields
			{
				Name:        "assignee_account_id",
				Description: "Account Id the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Assignee.AccountID"),
			},
			{
				Name:        "assignee_display_name",
				Description: "Display name the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Assignee.DisplayName"),
			},
			{
				Name:        "creator_account_id",
				Description: "Account Id of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Creator.AccountID"),
			},
			{
				Name:        "creator_display_name",
				Description: "Display name of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Creator.DisplayName"),
			},
			{
				Name:        "created",
				Description: "Time when the issue was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("V3Issue.Fields.Created").Transform(convertJiraTime),
			},
			{
				Name:        "duedate",
				Description: "Time by which the issue is expected to be completed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("V3Issue.Fields.DueDate").NullIfZero().Transform(convertJiraDate),
			},
			{
				Name:        "description",
				Description: "Description of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Description").Transform(extractDescription),
			},
			{
				Name:        "type",
				Description: "The name of the issue type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(getTypeFromFields),
			},
			{
				Name:        "labels",
				Description: "A list of labels applied to the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("V3Issue.Fields.Labels"),
			},
			{
				Name:        "priority",
				Description: "Priority assigned to the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Priority.Name"),
			},
			{
				Name:        "project_name",
				Description: "Name of the project to that issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Project.Name"),
			},
			{
				Name:        "reporter_account_id",
				Description: "Account Id of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Reporter.AccountID"),
			},
			{
				Name:        "reporter_display_name",
				Description: "Display name of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Reporter.DisplayName"),
			},
			{
				Name:        "resolution_date",
				Description: "Date the issue was resolved.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("V3Issue.Fields.ResolutionDate").NullIfZero().Transform(convertJiraTime),
			},
			{
				Name:        "summary",
				Description: "Details of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("V3Issue.Fields.Summary"),
			},
			{
				Name:        "updated",
				Description: "Time when the issue was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("V3Issue.Fields.Updated").Transform(convertJiraTime),
			},

			// JSON fields
			{
				Name:        "components",
				Description: "List of components associated with the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("V3Issue.Fields.Components").Transform(extractComponentIds),
			},
			{
				Name:        "fields",
				Description: "Json object containing important subfields of the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("V3Issue.Fields"),
			},
			{
				Name:        "changelog",
				Description: "JSON object containing changelog of the issue.",
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
		}),
	}
}

//// LIST FUNCTION

func listIssues(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	project := h.Item.(Project)

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var limit int = 10
	if d.QueryContext.Limit != nil {
		if *queryLimit < 10 {
			limit = int(*queryLimit)
		}
	}

	projectId := d.EqualsQualString("project_id")
	projectName := d.EqualsQualString("project_name")
	projectKey := d.EqualsQualString("project_key")

	if projectId != "" && projectId != project.ID {
		return nil, nil
	}
	if projectName != "" && projectName != project.Name {
		return nil, nil
	}
	if projectKey != "" && projectKey != project.Key {
		return nil, nil
	}

	jql := ""

	// The "jira_issue_comment" table is a child of the "jira_issue" table. When querying the child table, the parent table is executed first.
	// This restriction is necessary to correctly build the input parameters when querying only the "jira_issue" table.
	// Without this check, a query like "select title, id, issue_id, body from jira_issue_comment where issue_id = '10015'" will fail with an error.
	// Error: jira: request failed. Please analyze the request body for more details. Status code: 400: could not parse JSON: unexpected end of JSON input
	if d.Table.Name == "jira_issue" {
		qualJQL := buildJQLQueryFromQuals(d.Quals, d.Table.Columns)

		// Always include project key to avoid unbounded JQL error
		if qualJQL == "" {
			jql = fmt.Sprintf("project=%s", project.Key)
		} else {
			jql = fmt.Sprintf("project=%s AND %s", project.Key, qualJQL)
		}
	}

	requestBody := map[string]interface{}{
		"jql":        jql,
		"maxResults": limit,
		"fields":     []string{"*all"},
		"expand":     "names,changelog",
	}

	for {
		searchResult, _, err := searchWithContext(ctx, d, requestBody)
		if searchResult.NextPageToken != "" {
			requestBody["nextPageToken"] = searchResult.NextPageToken
		}
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue.listIssues", "search_error", err)
			return nil, err
		}
		issues := searchResult.Issues

		if err != nil {
			if isNotFoundError(err) || isBadRequestError(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("jira_issue.listIssues", "api_error", err)
			return nil, err
		}

		for _, issue := range issues {
			// Note: v3 API doesn't provide names in the response like v2 did
			// We'll need to handle epic and sprint fields differently or use hardcoded field names
			keys := map[string]string{
				"epic":   "customfield_10014",
				"sprint": "customfield_10020",
			}
			d.StreamListItem(ctx, IssueInfo{issue, keys})
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		// For v3 API, use IsLast to determine if there are more pages
		if searchResult.IsLast {
			return nil, nil
		}

	}
}

//// HYDRATE FUNCTION

func getIssue(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get the issue identifier from quals
	issueId := d.EqualsQualString("id")
	key := d.EqualsQualString("key")

	// Determine which identifier to use
	var issueIdentifier string
	if issueId != "" {
		issueIdentifier = issueId
	} else if key != "" {
		issueIdentifier = key
	} else {
		return nil, nil
	}

	// Connect to Jira API
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.getIssue", "connection_error", err)
		return nil, err
	}

	// Make V3 API request to get single issue
	req, err := client.NewRequestWithContext(ctx, "GET", fmt.Sprintf("rest/api/3/issue/%s", issueIdentifier), nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.getIssue", "request_creation_error", err)
		return nil, err
	}

	// Add query parameters for field expansion
	q := req.URL.Query()
	q.Add("expand", "names,changelog")
	q.Add("fields", "*all")
	req.URL.RawQuery = q.Encode()

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Execute the request
	var v3Issue V3Issue
	resp, err := client.Do(req, &v3Issue)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			// Issue not found - return nil (not an error for Get operations)
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_issue.getIssue", "api_error", err)
		return nil, err
	}

	// Create field keys mapping for custom fields
	keys := map[string]string{
		"epic":   "customfield_10014",
		"sprint": "customfield_10020",
	}

	// Return the issue info
	return IssueInfo{
		V3Issue: v3Issue,
		Keys:    keys,
	}, nil
}

func getStatusValue(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	issue := h.Item.(IssueInfo)
	issueStauus := d.EqualsQualString("status")

	if issueStauus != "" {
		return issueStauus, nil
	}
	return issue.V3Issue.Fields.Status.Name, nil
}

//// TRANSFORM FUNCTION

func extractComponentIds(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var componentIds []string
	for _, item := range d.Value.([]V3Component) {
		componentIds = append(componentIds, item.ID)
	}
	return componentIds, nil
}

func extractRequiredField(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issueInfo := d.HydrateItem.(IssueInfo)
	param := d.Param.(string)

	// For epic, check if there's a parent that is an Epic
	if param == "epic" {
		if issueInfo.V3Issue.Fields.Parent != nil {
			// Check if parent's issue type is Epic
			if issueInfo.V3Issue.Fields.Parent.Fields.IssueType.Name == "Epic" {
				return issueInfo.V3Issue.Fields.Parent.Key, nil
			}
		}
	}

	// Fallback: try to get the custom field key for the requested field type
	if fieldKey, exists := issueInfo.Keys[param]; exists {
		// Access the custom field from the raw JSON fields
		if fieldsBytes, err := json.Marshal(issueInfo.V3Issue.Fields); err == nil {
			var fieldsMap map[string]interface{}
			if err := json.Unmarshal(fieldsBytes, &fieldsMap); err == nil {
				if value, exists := fieldsMap[fieldKey]; exists {
					return value, nil
				}
			}
		}
	}
	return nil, nil
}

func extractSprintIds(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	// Handle both []V3Sprint (from struct) and []interface{} (from raw JSON)
	if sprints, ok := d.Value.([]V3Sprint); ok {
		var sprintIds []int
		for _, sprint := range sprints {
			sprintIds = append(sprintIds, sprint.ID)
		}
		return sprintIds, nil
	}

	// Fallback for []interface{} (raw JSON)
	if items, ok := d.Value.([]interface{}); ok {
		var sprintIds []interface{}
		for _, item := range items {
			if sprint, ok := item.(map[string]interface{}); ok {
				sprintIds = append(sprintIds, sprint["id"])
			}
		}
		return sprintIds, nil
	}

	return nil, nil
}

func extractSprintNames(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	// Handle both []V3Sprint (from struct) and []interface{} (from raw JSON)
	if sprints, ok := d.Value.([]V3Sprint); ok {
		var sprintNames []string
		for _, sprint := range sprints {
			sprintNames = append(sprintNames, sprint.Name)
		}
		return sprintNames, nil
	}

	// Fallback for []interface{} (raw JSON)
	if items, ok := d.Value.([]interface{}); ok {
		var sprintNames []interface{}
		for _, item := range items {
			if sprint, ok := item.(map[string]interface{}); ok {
				sprintNames = append(sprintNames, sprint["name"])
			}
		}
		return sprintNames, nil
	}

	return nil, nil
}

func getIssueTags(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issue := d.HydrateItem.(IssueInfo)

	tags := make(map[string]bool)
	if len(issue.V3Issue.Fields.Labels) > 0 {
		for _, i := range issue.V3Issue.Fields.Labels {
			tags[i] = true
		}
	}
	return tags, nil
}

func extractText(node interface{}, texts *[]string) {
	switch v := node.(type) {
	case map[string]interface{}:
		// If there's a "text" field, capture it
		if t, ok := v["text"].(string); ok {
			*texts = append(*texts, t)
		}
		// If there's a "content" array, recurse
		if arr, ok := v["content"].([]interface{}); ok {
			for _, item := range arr {
				extractText(item, texts)
			}
		}
	case []interface{}:
		for _, item := range v {
			extractText(item, texts)
		}
	}
}

func extractDescription(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	var data map[string]interface{}
	b, err := json.Marshal(d.Value)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	texts := []string{}
	extractText(data, &texts)

	// Join with newline
	result := strings.Join(texts, "\n")

	// Fallback to string representation
	return result, nil
}

func extractIssueTypeName(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	fieldsMap, ok := d.Value.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	issuetype, exists := fieldsMap["issuetype"]
	if !exists {
		return nil, nil
	}

	issuetypeMap, ok := issuetype.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	name, exists := issuetypeMap["name"]
	if !exists {
		return nil, nil
	}

	return name, nil
}

func getTypeFromFields(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issueInfo := d.HydrateItem.(IssueInfo)

	// Get the V3Issue JSON directly and access the fields
	if issueInfo.V3Issue.Fields.IssueType.Name != "" {
		return issueInfo.V3Issue.Fields.IssueType.Name, nil
	}

	return nil, nil
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

func searchWithContext(ctx context.Context, d *plugin.QueryData, requestBody map[string]interface{}) (*searchResult, *jira.Response, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.listIssues.searchWithContext", "connection_error", err)
		return nil, nil, err
	}

	// Create POST request
	req, err := client.NewRequestWithContext(ctx, "POST", "rest/api/3/search/jql", requestBody)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.searchWithContext", "request_creation_error", err)
		return nil, nil, err
	}

	// Set content type for JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	v := new(searchResult)
	resp, err := client.Do(req, v)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue.searchWithContext", "api_request_error", err)
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
	V3Issue
	Keys map[string]string
}

// V3 API Response Structures
type searchResult struct {
	Issues        []V3Issue `json:"issues"`
	IsLast        bool      `json:"isLast"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

type V3Issue struct {
	Expand string   `json:"expand"`
	ID     string   `json:"id"`
	Self   string   `json:"self"`
	Key    string   `json:"key"`
	Fields V3Fields `json:"fields"`
}

type V3Fields struct {
	Summary               string           `json:"summary"`
	Created               string           `json:"created"`
	Updated               string           `json:"updated"`
	Description           interface{}      `json:"description"` // Can be complex object or string
	IssueType             V3IssueType      `json:"issuetype"`
	Project               V3Project        `json:"project"`
	Reporter              V3User           `json:"reporter"`
	Creator               V3User           `json:"creator"`
	Assignee              *V3User          `json:"assignee"` // Can be null
	Priority              *V3Priority      `json:"priority"` // Can be null
	Status                V3Status         `json:"status"`
	StatusCategory        V3StatusCategory `json:"statusCategory"`
	Resolution            *V3Resolution    `json:"resolution"`     // Can be null
	ResolutionDate        *string          `json:"resolutiondate"` // Can be null
	DueDate               *string          `json:"duedate"`        // Can be null
	Components            []V3Component    `json:"components"`
	Labels                []string         `json:"labels"`
	FixVersions           []V3Version      `json:"fixVersions"`
	Versions              []V3Version      `json:"versions"`
	Attachment            []V3Attachment   `json:"attachment"`
	SubTasks              []V3SubTask      `json:"subtasks"`
	IssueLinks            []V3IssueLink    `json:"issuelinks"`
	Worklog               V3Worklog        `json:"worklog"`
	TimeTracking          V3TimeTracking   `json:"timetracking"`
	TimeSpent             *int             `json:"timespent"`
	TimeOriginalEstimate  *int             `json:"timeoriginalestimate"`
	AggregateTimeSpent    *int             `json:"aggregatetimespent"`
	AggregateTimeEstimate *int             `json:"aggregatetimeestimate"`
	Watches               V3Watches        `json:"watches"`
	LastViewed            *string          `json:"lastViewed"`
	Parent                *V3Parent        `json:"parent"` // Can be null for top-level issues
	// Sprint custom field
	Sprints []V3Sprint `json:"customfield_10020"` // Sprint information
	// Custom fields (using interface{} to handle any custom field)
	CustomFields map[string]interface{} `json:"-"` // We'll populate this separately
}

type V3Parent struct {
	ID     string         `json:"id"`
	Key    string         `json:"key"`
	Self   string         `json:"self"`
	Fields V3ParentFields `json:"fields"`
}

type V3ParentFields struct {
	Summary   string      `json:"summary"`
	Status    V3Status    `json:"status"`
	Priority  V3Priority  `json:"priority"`
	IssueType V3IssueType `json:"issuetype"`
}

type V3IssueType struct {
	Self           string `json:"self"`
	ID             string `json:"id"`
	Description    string `json:"description"`
	IconURL        string `json:"iconUrl"`
	Name           string `json:"name"`
	SubTask        bool   `json:"subtask"`
	AvatarID       int    `json:"avatarId"`
	EntityID       string `json:"entityId"`
	HierarchyLevel int    `json:"hierarchyLevel"`
}

type V3Project struct {
	Self           string            `json:"self"`
	ID             string            `json:"id"`
	Key            string            `json:"key"`
	Name           string            `json:"name"`
	ProjectTypeKey string            `json:"projectTypeKey"`
	Simplified     bool              `json:"simplified"`
	AvatarUrls     map[string]string `json:"avatarUrls"`
}

type V3User struct {
	Self         string            `json:"self"`
	AccountID    string            `json:"accountId"`
	EmailAddress string            `json:"emailAddress,omitempty"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
	DisplayName  string            `json:"displayName"`
	Active       bool              `json:"active"`
	TimeZone     string            `json:"timeZone,omitempty"`
	AccountType  string            `json:"accountType"`
}

type V3Priority struct {
	Self    string `json:"self"`
	IconURL string `json:"iconUrl"`
	Name    string `json:"name"`
	ID      string `json:"id"`
}

type V3Status struct {
	Self           string           `json:"self"`
	Description    string           `json:"description"`
	IconURL        string           `json:"iconUrl"`
	Name           string           `json:"name"`
	ID             string           `json:"id"`
	StatusCategory V3StatusCategory `json:"statusCategory"`
}

type V3StatusCategory struct {
	Self      string `json:"self"`
	ID        int    `json:"id"`
	Key       string `json:"key"`
	ColorName string `json:"colorName"`
	Name      string `json:"name"`
}

type V3Resolution struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

type V3Component struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type V3Version struct {
	Self            string `json:"self"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Archived        bool   `json:"archived"`
	Released        bool   `json:"released"`
	StartDate       string `json:"startDate,omitempty"`
	ReleaseDate     string `json:"releaseDate,omitempty"`
	UserStartDate   string `json:"userStartDate,omitempty"`
	UserReleaseDate string `json:"userReleaseDate,omitempty"`
}

type V3Attachment struct {
	Self      string `json:"self"`
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Author    V3User `json:"author"`
	Created   string `json:"created"`
	Size      int    `json:"size"`
	MimeType  string `json:"mimeType"`
	Content   string `json:"content"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

type V3SubTask struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Self         string `json:"self"`
	OutwardIssue struct {
		ID     string `json:"id"`
		Key    string `json:"key"`
		Self   string `json:"self"`
		Fields struct {
			Status V3Status `json:"status"`
		} `json:"fields"`
	} `json:"outwardIssue"`
}

type V3IssueLink struct {
	ID           string          `json:"id"`
	Self         string          `json:"self"`
	Type         V3IssueLinkType `json:"type"`
	InwardIssue  *V3LinkedIssue  `json:"inwardIssue,omitempty"`
	OutwardIssue *V3LinkedIssue  `json:"outwardIssue,omitempty"`
}

type V3IssueLinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
	Self    string `json:"self"`
}

type V3LinkedIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Status    V3Status    `json:"status"`
		Priority  V3Priority  `json:"priority"`
		IssueType V3IssueType `json:"issuetype"`
		Summary   string      `json:"summary"`
	} `json:"fields"`
}

type V3Worklog struct {
	StartAt    int              `json:"startAt"`
	MaxResults int              `json:"maxResults"`
	Total      int              `json:"total"`
	Worklogs   []V3WorklogEntry `json:"worklogs"`
}

type V3WorklogEntry struct {
	Self             string `json:"self"`
	Author           V3User `json:"author"`
	Comment          string `json:"comment,omitempty"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
	Started          string `json:"started"`
	TimeSpent        string `json:"timeSpent"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
	ID               string `json:"id"`
	IssueID          string `json:"issueId"`
}

type V3TimeTracking struct {
	OriginalEstimate         string `json:"originalEstimate,omitempty"`
	RemainingEstimate        string `json:"remainingEstimate,omitempty"`
	TimeSpent                string `json:"timeSpent,omitempty"`
	OriginalEstimateSeconds  int    `json:"originalEstimateSeconds,omitempty"`
	RemainingEstimateSeconds int    `json:"remainingEstimateSeconds,omitempty"`
	TimeSpentSeconds         int    `json:"timeSpentSeconds,omitempty"`
}

type V3Watches struct {
	Self       string `json:"self"`
	WatchCount int    `json:"watchCount"`
	IsWatching bool   `json:"isWatching"`
}

type V3Sprint struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	State        string `json:"state"`
	BoardID      int    `json:"boardId,omitempty"`
	Goal         string `json:"goal,omitempty"`
	StartDate    string `json:"startDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	CompleteDate string `json:"completeDate,omitempty"`
}
