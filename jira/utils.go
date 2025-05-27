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

//// CONNECT FUNCTION

func connect(_ context.Context, d *plugin.QueryData) (*jira.Client, error) {
	cacheKey := "atlassian-jira"
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.(*jira.Client), nil
	}

	baseUrl := os.Getenv("JIRA_URL")
	username := os.Getenv("JIRA_USER")
	token := os.Getenv("JIRA_TOKEN")
	personalAccessToken := os.Getenv("JIRA_PERSONAL_ACCESS_TOKEN")

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
		personalAccessToken = *jiraConfig.PersonalAccessToken
	}

	if baseUrl == "" {
		return nil, errors.New("'base_url' must be set in the connection configuration")
	}
	if username == "" && token != "" {
		return nil, errors.New("'token' is set but 'username' is not set in the connection configuration")
	}
	if token == "" && personalAccessToken == "" {
		return nil, errors.New("'token' or 'personal_access_token' must be set in the connection configuration")
	}
	if token != "" && personalAccessToken != "" {
		return nil, errors.New("'token' and 'personal_access_token' are both set, please use only one auth method")
	}

	var client *jira.Client
	var err error

	if personalAccessToken != "" {
		tokenProvider := jirav2.BearerAuthTransport{Token: personalAccessToken}
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

	d.ConnectionManager.Cache.Set(cacheKey, client)

	return client, nil
}

//// Constants
const (
	ColumnDescriptionTitle = "Title of the resource."
)

//// TRANSFORM FUNCTIONS

func convertJiraTime(_ context.Context, d *transform.TransformData) (interface{}, error) {
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

func convertJiraDate(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	return time.Time(d.Value.(jira.Date)), nil
}

//// JQL Builder

func buildJQLQueryFromQuals(equalQuals plugin.KeyColumnQualMap, tableColumns []*plugin.Column) string {
	filters := []string{}

	for _, filterQualItem := range tableColumns {
		filterQual := equalQuals[filterQualItem.Name]
		if filterQual == nil {
			continue
		}

		if filterQual.Name == filterQualItem.Name {
			if filterQual.Quals == nil {
				continue
			}

			for _, qual := range filterQual.Quals {
				if qual.Value != nil {
					value := qual.Value
					switch filterQualItem.Type {
					case proto.ColumnType_STRING:
						switch qual.Operator {
						case "=":
							filters = append(filters, fmt.Sprintf("\"%s\" = \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
						case "<>":
							filters = append(filters, fmt.Sprintf("%s != \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
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

func getIssueJQLKey(columnName string) string {
	remappedColumns := map[string]string{
		"resolution_date": "resolutiondate",
		"status_category": "statuscategory",
	}

	if val, ok := remappedColumns[columnName]; ok {
		return val
	}
	return strings.ToLower(strings.Split(columnName, "_")[0])
}

//// WORKERS MANAGEMENT

func getMaxWorkers(ctx context.Context, d *plugin.QueryData) int {
	const defaultWorkers = 1
	const cacheKey = "jira_max_workers"

	if cached, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		if workers, ok := cached.(int); ok {
			return workers
		}
	}

	maxWorkers := defaultWorkers

	jiraConfig := GetConfig(d.Connection)

	if jiraConfig.Workers != nil && *jiraConfig.Workers > 0 {
		maxWorkers = *jiraConfig.Workers
	}

	d.ConnectionManager.Cache.Set(cacheKey, maxWorkers)

	plugin.Logger(ctx).Warn("getMaxWorkers", "found_workers", maxWorkers)

	return maxWorkers
}

//// DYNAMIC FIELDS BUILDER

func getRequestedFields(ctx context.Context, d *plugin.QueryData) []string {
	if d.QueryContext.Columns == nil || len(d.QueryContext.Columns) == 0 {
		return nil
	}

	fieldMap := map[string]string{
		"id":                    "id",
		"key":                   "key",
		"summary":               "summary",
		"status":                "status",
		"status_category":       "status",
		"project_key":           "project",
		"project_id":            "project",
		"project_name":          "project",
		"created":               "created",
		"updated":               "updated",
		"duedate":               "duedate",
		"assignee_account_id":   "assignee",
		"assignee_display_name": "assignee",
		"creator_account_id":    "creator",
		"creator_display_name":  "creator",
		"reporter_account_id":   "reporter",
		"reporter_display_name": "reporter",
		"resolution_date":       "resolutiondate",
		"priority":              "priority",
		"type":                  "issuetype",
		"labels":                "labels",
		"components":            "components",
		"fields":                "*",
	}

	fields := make(map[string]bool)

	for _, column := range d.QueryContext.Columns {
		jiraField, exists := fieldMap[column]
		if exists {
			fields[jiraField] = true
		} else {
			fields["*"] = true
		}
	}

	if fields["*"] {
		return nil
	}

	var fieldList []string
	for f := range fields {
		fieldList = append(fieldList, f)
	}

	return fieldList
}