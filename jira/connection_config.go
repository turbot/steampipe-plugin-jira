package jira

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type jiraConfig struct {
	BaseUrl             *string `cty:"base_url"`
	Username            *string `cty:"username"`
	Token               *string `cty:"token"`
	PersonalAccessToken *string `cty:"personal_access_token"`
	PageSize            *int    `cty:"page_size"`
	CaseSensitivity     *string `cty:"case_sensitivity"`
	IssueLimit          *int    `cty:"issue_limit"`
	ComponentLimit      *int    `cty:"component_limit"`
	ProjectLimit        *int    `cty:"project_limit"`
	BoardLimit          *int    `cty:"board_limit"`
	SprintLimit         *int    `cty:"sprint_limit"`
	RowLimitError       *bool   `cty:"row_limit_error"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"base_url": {
		Type: schema.TypeString,
	},
	"username": {
		Type: schema.TypeString,
	},
	"token": {
		Type: schema.TypeString,
	},
	"personal_access_token": {
		Type: schema.TypeString,
	},
	"page_size": {
		Type: schema.TypeInt,
	},
	"case_sensitivity": {
		Type: schema.TypeString,
	},
	"issue_limit": {
		Type: schema.TypeInt,
	},
	"component_limit": {
		Type: schema.TypeInt,
	},
	"project_limit": {
		Type: schema.TypeInt,
	},
	"board_limit": {
		Type: schema.TypeInt,
	},
	"sprint_limit": {
		Type: schema.TypeInt,
	},
	"row_limit_error": {
		Type: schema.TypeBool,
	},
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
