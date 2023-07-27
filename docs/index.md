---
organization: Turbot
category: ["software development"]
icon_url: "/images/plugins/turbot/jira.svg"
brand_color: "#2684FF"
display_name: "Jira"
short_name: "jira"
description: "Steampipe plugin for querying sprints, issues, epics and more from Jira."
og_description: "Query Jira with SQL! Open source CLI. No DB required."
og_image: "/images/plugins/turbot/jira-social-graphic.png"
---

# Jira + Steampipe

[Jira](https://www.atlassian.com/software/jira) provides on-demand cloud computing platforms and APIs to plan,
track, and release great software.

[Steampipe](https://steampipe.io) is an open source CLI to instantly query cloud APIs using SQL.

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

## Documentation

- **[Table definitions & examples →](/plugins/turbot/jira/tables)**

## Get started

### Install

Download and install the latest Jira plugin:

```bash
steampipe plugin install jira
```

### Credentials

| Item        | Description                                                                                                                            |
| :---------- | :------------------------------------------------------------------------------------------------------------------------------------- |
| Credentials | Jira requires an [API token](https://id.atlassian.com/manage-profile/security/api-tokens), sit base url and username for all requests. |
| Radius      | Each connection represents a single Jira site.                                                                                         |

<!-- | Permissions | Grant the `ReadOnlyAccess` policy to your user or role.                                                                                | -->

### Configuration

Installing the latest jira plugin will create a config file (`~/.steampipe/config/jira.spc`) with a single connection named `jira`:

```hcl
connection "jira" {
  plugin = "jira"

  # The baseUrl of your Jira Instance API
  # Can also be set with the JIRA_URL environment variable.
  # base_url = "https://your-domain.atlassian.net/"

  # The user name to access the jira cloud instance
  # Can also be set with the `JIRA_USER` environment variable.
  # username = "abcd@xyz.com"

  # Access Token for which to use for the API
  # Can also be set with the `JIRA_TOKEN` environment variable.
  # token = "8WqcdT0rvIZpCjtDqReF48B1"
}
```

- `base_url` - The site url of your attlassian jira subscription.
- `username` - Email address of agent user who have permission to access the API.
- `token` - [API token](https://id.atlassian.com/manage-profile/security/api-tokens) for user's Atlassian account.

Alternatively, you can also use the standard Jira environment variables to obtain credentials **only if other arguments (`base_url`, `username` and `token`) are not specified** in the connection:

```sh
export JIRA_URL=https://your-domain.atlassian.net/
export JIRA_USER=abcd@xyz.com
export JIRA_TOKEN=8WqcdT0rvIZpCjtDqReF48B1
```

## Get involved

- Open source: https://github.com/turbot/steampipe-plugin-jira
- Community: [Join #steampipe on Slack →](https://turbot.com/community/join)
