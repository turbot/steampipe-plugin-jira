package jira

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type jiraConfig struct {
	BaseUrl             *string `hcl:"base_url"`
	Username            *string `hcl:"username"`
	Token               *string `hcl:"token"`
	PersonalAccessToken *string `hcl:"personal_access_token"`
	PageSize            *int    `cty:"page_size"`
	CaseSensitivity     *string `cty:"case_sensitivity"`
}

func ConfigInstance() interface{} {
	return &jiraConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) jiraConfig {
	if connection == nil || connection.Config == nil {
		return jiraConfig{}
	}
	config, _ := connection.Config.(jiraConfig)
	return config
}
