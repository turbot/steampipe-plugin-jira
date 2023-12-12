![image](https://hub.steampipe.io/images/plugins/turbot/jira-social-graphic.png)

# Jira Plugin for Steampipe

Use SQL to query infrastructure including servers, networks, facilities and more from Jira.

- **[Get started →](https://hub.steampipe.io/plugins/turbot/jira)**
- Documentation: [Table definitions & examples](https://hub.steampipe.io/plugins/turbot/jira/tables)
- Community: [Join #steampipe on Slack →](https://turbot.com/community/join)
- Get involved: [Issues](https://github.com/turbot/steampipe-plugin-jira/issues)

## Quick start

### Install

Download and install the latest Jira plugin:

```bash
steampipe plugin install jira
```

Configure your [credentials](https://hub.steampipe.io/plugins/turbot/jira#credentials) and [config file](https://hub.steampipe.io/plugins/turbot/jira#configuration).

Configure your account details in `~/.steampipe/config/jira.spc`:

```hcl
connection "jira" {
  plugin = jira

  # Authentication information
  base_url              = "https://your-domain.atlassian.net/"
  username              = "abcd@xyz.com"
  token                 = "8WqcdT0rvIZpCjtDqReF48B1"
}
```

For [self-hosted Jira instances](https://github.com/andygrunwald/go-jira/#bearer---personal-access-tokens-self-hosted-jira), please use the `personal_access_token` field instead of `token`. This access token can only be used to query `jira_backlog_issue`, `jira_board`, `jira_issue` and `jira_sprint` tables.

```hcl
connection "jira" {
  plugin = jira

  # Authentication information
  base_url              = "https://your-domain.atlassian.net/"
  personal_access_token = "MDU0MDMx7cE25TQ3OujDfy/vkv/eeSXXoh/zXY1ex9cp"
}
```

Or through environment variables:

```sh
export JIRA_URL=https://your-domain.atlassian.net/
export JIRA_USER=abcd@xyz.com
export JIRA_TOKEN=8WqcdT0rvIZpCjtDqReF48B1
export JIRA_PERSONAL_ACCESS_TOKEN="MDU0MDMx7cE25TQ3OujDfy/vkv/eeSXXoh/zXY1ex9cp"
```

Run steampipe:

```shell
steampipe query
```

List users in your Jira account:

```sql
select
  display_name,
  account_type as type,
  active as status,
  account_id
from
  jira_user;
```

```
+-------------------------------+-----------+--------+-----------------------------+
| display_name                  | type      | status | account_id                  |
+-------------------------------+-----------+--------+-----------------------------+
| Confluence Analytics (System) | app       | true   | 557058:cbc04d7be567aa5332c6 |
| John Smyth                    | atlassian | true   | 1f2e1d34e0e56a001ea44fc1    |
+-------------------------------+-----------+--------+-----------------------------+
```

## Engines

This plugin is available for the following engines:

| Engine        | Description
|---------------|------------------------------------------
| [Steampipe](https://steampipe.io/docs) | The Steampipe CLI exposes APIs and services as a high-performance relational database, giving you the ability to write SQL-based queries to explore dynamic data. Mods extend Steampipe's capabilities with dashboards, reports, and controls built with simple HCL. The Steampipe CLI is a turnkey solution that includes its own Postgres database, plugin management, and mod support.
| [Postgres FDW](https://steampipe.io/docs/steampipe_postgres/index) | Steampipe Postgres FDWs are native Postgres Foreign Data Wrappers that translate APIs to foreign tables. Unlike Steampipe CLI, which ships with its own Postgres server instance, the Steampipe Postgres FDWs can be installed in any supported Postgres database version.
| [SQLite Extension](https://steampipe.io/docs//steampipe_sqlite/index) | Steampipe SQLite Extensions provide SQLite virtual tables that translate your queries into API calls, transparently fetching information from your API or service as you request it.
| [Export](https://steampipe.io/docs/steampipe_export/index) | Steampipe Plugin Exporters provide a flexible mechanism for exporting information from cloud services and APIs. Each exporter is a stand-alone binary that allows you to extract data using Steampipe plugins without a database.
| [Turbot Pipes](https://turbot.com/pipes/docs) | Turbot Pipes is the only intelligence, automation & security platform built specifically for DevOps. Pipes provide hosted Steampipe database instances, shared dashboards, snapshots, and more.

## Developing

Prerequisites:

- [Steampipe](https://steampipe.io/downloads)
- [Golang](https://golang.org/doc/install)

Clone:

```sh
git clone https://github.com/turbot/steampipe-plugin-jira.git
cd steampipe-plugin-jira
```

Build, which automatically installs the new version to your `~/.steampipe/plugins` directory:

```
make
```

Configure the plugin:

```
cp config/* ~/.steampipe/config
vi ~/.steampipe/config/jira.spc
```

Try it!

```
steampipe query
> .inspect jira
```

Further reading:

- [Writing plugins](https://steampipe.io/docs/develop/writing-plugins)
- [Writing your first table](https://steampipe.io/docs/develop/writing-your-first-table)

## Open Source & Contributing

This repository is published under the [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0) (source code) and [CC BY-NC-ND](https://creativecommons.org/licenses/by-nc-nd/2.0/) (docs) licenses. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). We look forward to collaborating with you!

[Steampipe](https://steampipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).

## Get Involved

**[Join #steampipe on Slack →](https://turbot.com/community/join)**

Want to help but don't know where to start? Pick up one of the `help wanted` issues:

- [Steampipe](https://github.com/turbot/steampipe/labels/help%20wanted)
- [Jira Plugin](https://github.com/turbot/steampipe-plugin-jira/labels/help%20wanted)
