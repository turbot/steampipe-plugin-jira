package main

import (
	"github.com/jdevoo/steampipe-plugin-jira/jira"

	"github.com/turbot/steampipe-plugin-sdk/v3/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: jira.Plugin})
}
