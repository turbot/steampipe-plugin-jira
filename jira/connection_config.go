package jira

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type jiraConfig struct {
	BaseUrl             *string `hcl:"base_url"`
	Username            *string `hcl:"username"`
	Token               *string `hcl:"token"`
	PersonalAccessToken *string `hcl:"personal_access_token"`
	PageSize            *int    `hcl:"page_size"`
	CaseSensitivity     *string `hcl:"case_sensitivity"`
	IssueLimit          *int    `hcl:"issue_limit"`
	ComponentLimit      *int    `hcl:"component_limit"`
	ProjectLimit        *int    `hcl:"project_limit"`
	BoardLimit          *int    `hcl:"board_limit"`
	SprintLimit         *int    `hcl:"sprint_limit"`
	RowLimitError       *bool   `hcl:"row_limit_error"`
	ClientId            *string `hcl:"client_id"`
	ClientSecret        *string `hcl:"client_secret"`
	RedirectUri         *string `hcl:"redirect_uri"`
	RefreshToken        *string `hcl:"refresh_token"`
	RefreshUri          *string `hcl:"refresh_uri"`
	AuthMode            *string `hcl:"auth_mode"`
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
