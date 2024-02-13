package jira

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	jirav2 "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func connect(_ context.Context, d *plugin.QueryData) (*jira.Client, error) {

	// Load connection from cache, which preserves throttling protection etc
	cacheKey := "atlassian-jira"
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.(*jira.Client), nil
	}

	// Default to the env var settings
	baseUrl := os.Getenv("JIRA_URL")
	username := os.Getenv("JIRA_USER")
	token := os.Getenv("JIRA_TOKEN")
	personal_access_token := os.Getenv("JIRA_PERSONAL_ACCESS_TOKEN")

	// Prefer config options given in Steampipe
	jiraConfig := GetConfig(d.Connection)

	if jiraConfig.BaseUrl != nil {
		baseUrl = *jiraConfig.BaseUrl
	}
	if jiraConfig.Username != nil {
		username = *jiraConfig.Username
	}
	if jiraConfig.Token != nil {
		token = *jiraConfig.Token
	}
	if jiraConfig.PersonalAccessToken != nil {
		personal_access_token = *jiraConfig.PersonalAccessToken
	}

	if baseUrl == "" {
		return nil, errors.New("'base_url' must be set in the connection configuration")
	}
	if username == "" && token != "" {
		return nil, errors.New("'token' is set but 'username' is not set in the connection configuration")
	}
	if token == "" && personal_access_token == "" {
		return nil, errors.New("'token' or 'personal_access_token' must be set in the connection configuration")
	}
	if token != "" && personal_access_token != "" {
		return nil, errors.New("'token' and 'personal_access_token' are both set, please use only one auth method")
	}

	var client *jira.Client
	var err error

	if personal_access_token != "" {
		// If the username is empty, let's assume the user is using a PAT
		tokenProvider := jirav2.BearerAuthTransport{Token: personal_access_token}
		client, err = jira.NewClient(tokenProvider.Client(), baseUrl)
	} else {
		tokenProvider := jira.BasicAuthTransport{
			Username: username,
			Password: token,
		}
		client, err = jira.NewClient(tokenProvider.Client(), baseUrl)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating Jira client: %s", err.Error())
	}

	// Save to cache
	d.ConnectionManager.Cache.Set(cacheKey, client)

	// Done
	return client, nil
}

// // Constants
const (
	ColumnDescriptionTitle = "Title of the resource."
)

//// TRANSFORM FUNCTION

// convertJiraTime:: converts jira.Time to time.Time
func convertJiraTime(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	if v, ok := d.Value.(jira.Time); ok {
		return time.Time(v), nil
	} else if v, ok := d.Value.(*jira.Time); ok {
		return time.Time(*v), nil
	}
	return nil, nil
}

// convertJiraDate:: converts jira.Date to time.Time
func convertJiraDate(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	return time.Time(d.Value.(jira.Date)), nil
}

func buildJQLQueryFromQuals(ctx context.Context, equalQuals plugin.KeyColumnQualMap, tableColumns []*plugin.Column) string {
	filters := []string{}
	plugin.Logger(ctx).Debug("jira_issue::buildJQLQueryFromQuals", equalQuals)
	for _, filterQualItem := range tableColumns {
		filterQual := equalQuals[filterQualItem.Name]

		if filterQual == nil {
			continue
		}

		// Check only if filter qual map matches with optional column name
		if filterQual.Name == filterQualItem.Name {
			if filterQual.Quals == nil {
				continue
			}

			for _, qual := range filterQual.Quals {
				if qual.Value != nil {
					value := qual.Value
					//plugin.Logger(ctx).Debug("jira_issue::buildJQLQueryFromQuals", value, filterQualItem.Type, qual.Operator)

					switch filterQualItem.Type {
					case proto.ColumnType_STRING:
						switch qual.Operator {
						case "=":
							filters = append(filters, fmt.Sprintf("\"%s\" = \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
						case "<>":
							filters = append(filters, fmt.Sprintf("\"%s\" != \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
							//case "~~":
							//	filters = append(filters, fmt.Sprintf("%s = \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
						}
					case proto.ColumnType_TIMESTAMP:
						switch qual.Operator {
						case "=", ">=", ">", "<=", "<":
							filters = append(filters, fmt.Sprintf("\"%s\" %s \"%s\"", getIssueJQLKey(filterQualItem.Name), qual.Operator, value.GetTimestampValue().AsTime().Format("2006-01-02 15:04")))
						case "<>":
							filters = append(filters, fmt.Sprintf("\"%s\" != \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetTimestampValue().AsTime().Format("2006-01-02 15:04")))
						}

					}
				}
			}

		}
	}

	if len(filters) > 0 {
		return strings.Join(filters, " AND ")
	}

	return ""
}

func getRequiredCustomField() map[string]map[string]interface{} {
	customFieldMap := map[string]map[string]interface{}{
		"etv": {
			"key":        "customfield_13193",
			"name":       "Eng Target Version/s",
			"searchable": true,
			"type":       "array",
		},
		"release_commit": {
			"key":        "customfield_13139",
			"name":       "Release Commit",
			"searchable": true,
			"type":       "option",
		},
		"vteam": {
			"key":        "customfield_13323",
			"name":       "V-team/P-team",
			"searchable": true,
			"type":       "option-with-child",
		},
	}
	return customFieldMap
}

func getIssueJQLKey(columnName string) string {
	customFieldMap := getRequiredCustomField()
	jqlFieldMap := map[string]string{
		"resolution_date":   "resolutionDate",
		"status_category":   "statusCategory",
		"parent_key":        "parent",
		"parent_status":     "parentStatus",
		"parent_issue_type": "parentIssueType",
	}
	// if the column name is in the map, return the value else return the column name
	if field, ok := jqlFieldMap[columnName]; ok {
		return field
	} else if customField, ok := customFieldMap[columnName]; ok {
		return customField["name"].(string)
	} else {
		return strings.ToLower(strings.Split(columnName, "_")[0])
	}
}

func getPageSize(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	pageSize := 50
	if jiraConfig.PageSize != nil {
		pageSize = *jiraConfig.PageSize
	}

	if pageSize < 1 || pageSize > 100 {
		return -1, errors.New("'page_size' must be set to 1 to 100 in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	return pageSize, nil
}

func getCaseSensitivity(_ context.Context, d *plugin.QueryData) (string, error) {
	jiraConfig := GetConfig(d.Connection)

	caseSensitivity := "insensitive"
	if jiraConfig.CaseSensitivity != nil {
		caseSensitivity = *jiraConfig.CaseSensitivity
	}

	if caseSensitivity != "sensitive" && caseSensitivity != "insensitive" {
		return "", errors.New("'case_sensitive' must be set to 'insensitive' or 'sensitive' in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	return caseSensitivity, nil
}

func getIssueLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	issueLimit := 500
	if jiraConfig.IssueLimit != nil {
		issueLimit = *jiraConfig.IssueLimit
	}

	if issueLimit < 1 {
		return -1, errors.New("'issue_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return issueLimit, nil
}

func getComponentLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	componentLimit := 200
	if jiraConfig.ComponentLimit != nil {
		componentLimit = *jiraConfig.ComponentLimit
	}

	if componentLimit < 1 {
		return -1, errors.New("'component_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return componentLimit, nil
}

func getProjectLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	projectLimit := 200
	if jiraConfig.ProjectLimit != nil {
		projectLimit = *jiraConfig.ProjectLimit
	}

	if projectLimit < 1 {
		return -1, errors.New("'project_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return projectLimit, nil
}

func getBoardLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	boardLimit := 300
	if jiraConfig.BoardLimit != nil {
		boardLimit = *jiraConfig.BoardLimit
	}

	if boardLimit < 1 {
		return -1, errors.New("'board_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return boardLimit, nil
}

func getSprintLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	sprintLimit := 25
	if jiraConfig.SprintLimit != nil {
		sprintLimit = *jiraConfig.SprintLimit
	}

	if sprintLimit < 1 {
		return -1, errors.New("'sprint_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return sprintLimit, nil
}

func getRowLimitError(_ context.Context, d *plugin.QueryData) (bool, error) {
	jiraConfig := GetConfig(d.Connection)

	rowLimitError := true
	if jiraConfig.RowLimitError != nil {
		rowLimitError = *jiraConfig.RowLimitError
	}

	return rowLimitError, nil
}
