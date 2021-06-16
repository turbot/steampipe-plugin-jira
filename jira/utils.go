package jira

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

func connect(_ context.Context, d *plugin.QueryData) (*jira.Client, error) {

	// Load connection from cache, which preserves throttling protection etc
	cacheKey := "atlassian-jira"
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.(*jira.Client), nil
	}

	// Start with an empty Turbot config
	tokenProvider := jira.BasicAuthTransport{}
	var baseUrl string

	// Prefer config options given in Steampipe
	jiraConfig := GetConfig(d.Connection)

	if jiraConfig.BaseUrl != nil {
		baseUrl = *jiraConfig.BaseUrl
	}
	if jiraConfig.Username != nil {
		tokenProvider.Username = *jiraConfig.Username
	}
	if jiraConfig.Token != nil {
		tokenProvider.Password = *jiraConfig.Token
	}

	// Create the client
	client, err := jira.NewClient(tokenProvider.Client(), baseUrl)
	if err != nil {
		return nil, fmt.Errorf("error creating Jira client: %s", err.Error())
	}

	// Save to cache
	d.ConnectionManager.Cache.Set(cacheKey, client)

	// Done
	return client, nil
}

//// Constants
const (
	ColumnDescriptionTitle = "Title of the resource."
)

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "404")
}

//// TRANSFORM FUNCTION

// convertJiraTime:: converts jira.Time to time.Time
func convertJiraTime(_ context.Context, d *transform.TransformData) (interface{}, error) {
	return time.Time(d.Value.(jira.Time)), nil
}

// convertJiraDate:: converts jira.Date to time.Time
func convertJiraDate(_ context.Context, d *transform.TransformData) (interface{}, error) {
	return time.Time(d.Value.(jira.Date)), nil
}
