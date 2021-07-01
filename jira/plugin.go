package jira

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

const pluginName = "steampipe-plugin-jira"

// Plugin creates this (azure) plugin
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             pluginName,
		DefaultTransform: transform.FromCamel(),
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			Schema:      ConfigSchema,
		},
		TableMap: map[string]*plugin.Table{
			"jira_backlog_issue": tableBacklogIssue(ctx),
			"jira_board":         tableBoard(ctx),
			"jira_component":     tableComponent(ctx),
			"jira_dashboard":     tableDashboard(ctx),
			"jira_epic":          tableEpic(ctx),
			"jira_group":         tableGroup(ctx),
			"jira_issue":         tableIssue(ctx),
			"jira_priority":      tablePriority(ctx),
			"jira_project":       tableProject(ctx),
			"jira_project_role":  tableProjectRole(ctx),
			"jira_sprint":        tableSprint(ctx),
			"jira_user":          tableUser(ctx),
		},
	}

	return p
}
