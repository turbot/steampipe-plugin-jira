package main

import (
	"github.com/turbot/steampipe-plugin-jira/jira"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: jira.Plugin})
}
