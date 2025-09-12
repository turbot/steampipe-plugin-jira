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
func convertJiraTime(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	if v, ok := d.Value.(jira.Time); ok {
		return time.Time(v), nil
	} else if v, ok := d.Value.(*jira.Time); ok {
		return time.Time(*v), nil
	} else if v, ok := d.Value.(*string); ok {
		// Handle *string type from V3 API
		if v == nil || *v == "" {
			return nil, nil
		}
		layout := "2006-01-02T15:04:05.000-0700"
		return time.Parse(layout, *v)
	} else if v, ok := d.Value.(string); ok {
		// Handle empty strings from V3 API
		if v == "" {
			return nil, nil
		}
		layout := "2006-01-02T15:04:05.000-0700"
		return time.Parse(layout, v)
	}
	return nil, nil
}

// convertJiraDate:: converts jira.Date to time.Time
func convertJiraDate(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	switch t := d.Value.(type) {
	case jira.Date:
		return time.Time(t), nil
	case *jira.Date:
		return time.Time(*t), nil
	case *string:
		return time.Parse(time.DateOnly, *t)
	case string:
		return time.Parse(time.DateOnly, t)
	}
	return time.Time(d.Value.(jira.Date)), nil
}

func buildJQLQueryFromQuals(equalQuals plugin.KeyColumnQualMap, tableColumns []*plugin.Column) string {
	filters := []string{}
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
					switch filterQualItem.Type {
					case proto.ColumnType_STRING:
						jqlFieldName := getIssueJQLKey(filterQualItem.Name)
						switch qual.Operator {
						case "=":
							// Special handling for key field - don't quote field name but quote value
							if filterQualItem.Name == "key" {
								filters = append(filters, fmt.Sprintf("%s = \"%s\"", jqlFieldName, value.GetStringValue()))
							} else {
								filters = append(filters, fmt.Sprintf("\"%s\" = \"%s\"", jqlFieldName, value.GetStringValue()))
							}
						case "<>":
							// Special handling for key field
							if filterQualItem.Name == "key" {
								filters = append(filters, fmt.Sprintf("%s != \"%s\"", jqlFieldName, value.GetStringValue()))
							} else {
								filters = append(filters, fmt.Sprintf("%s != \"%s\"", jqlFieldName, value.GetStringValue()))
							}
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
		"key":             "key",
		"resolution_date": "resolutiondate",
		"status_category": "statuscategory",
	}

	if val, ok := remappedColumns[columnName]; ok {
		return val
	}
	return strings.ToLower(strings.Split(columnName, "_")[0])
}
