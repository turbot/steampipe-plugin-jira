package main

import (
	"github.com/turbot/steampipe-plugin-jira/jira"

	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: jira.Plugin})
}
