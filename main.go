package main

import (
	"github.com/turbot/steampipe-plugin-jira/v2/jira"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: jira.Plugin})
}
