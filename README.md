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

For [self-hosted Jira instances](https://github.com/andygrunwald/go-jira/#bearer---personal-access-tokens-self-hosted-jira), please use the `personal_access_token` field instead of `token`. This access token can only be used to query `jira_backlog_issue`, `jira_board`, `jira_issue` and `jira_sprint` tables.

Configure your account details in `~/.steampipe/config/jira.spc`:

```hcl
connection "jira" {
  plugin = jira

  # Authentication information
  base_url              = "https://your-domain.atlassian.net/"
  username              = "abcd@xyz.com"
  token                 = "8WqcdT0rvIZpCjtDqReF48B1"
  personal_access_token = "MDU0MDMx7cE25TQ3OujDfy/vkv/eeSXXoh/zXY1ex9cp"
}
```

For self-hosted Jira instances:

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

## Contributing

Please see the [contribution guidelines](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md) and our [code of conduct](https://github.com/turbot/steampipe/blob/main/CODE_OF_CONDUCT.md). All contributions are subject to the [Apache 2.0 open source license](https://github.com/turbot/steampipe-plugin-jira/blob/main/LICENSE).

`help wanted` issues:

- [Steampipe](https://github.com/turbot/steampipe/labels/help%20wanted)
- [Jira Plugin](https://github.com/turbot/steampipe-plugin-jira/labels/help%20wanted)
