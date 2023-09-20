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
  # You should leave it empty if you are using a Personal Access Token (PAT)
  # token = "8WqcdT0rvIZpCjtDqReF48B1"

  # Personal Access Token for which to use for the API.
  # This one isused in self-hosted Jira instances.
  # Can also be set with the `JIRA_PERSONAL_ACCESS_TOKEN` environment variable.
  # personal_access_token = "MDU0MDMx7cE25TQ3OujDfy/vkv/eeSXXoh/zXY1ex9cp"
}
```

- `base_url` - The site url of your attlassian jira subscription.
- `username` - Email address of agent user who have permission to access the API.
- `token` - [API token](https://id.atlassian.com/manage-profile/security/api-tokens) for user's Atlassian account.
- `personal_access_token` - [API PAT](https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html) for self hosted Jira instances.

Alternatively, you can also use the standard Jira environment variables to obtain credentials **only if other arguments (`base_url`, `username` and `token`) are not specified** in the connection:

```sh
export JIRA_URL=https://your-domain.atlassian.net/
export JIRA_USER=abcd@xyz.com
export JIRA_TOKEN=8WqcdT0rvIZpCjtDqReF48B1
export JIRA_PERSONAL_ACCESS_TOKEN="MDU0MDMx7cE25TQ3OujDfy/vkv/eeSXXoh/zXY1ex9cp"
```

## Important note about self hosted Jira instances

As reported in [this GH issue](https://github.com/turbot/steampipe-plugin-jira/pull/86#issuecomment-1697122416), there are some tables that don't work on self hosted Jira.

| table                 | works |
|-----------------------|:-----:|
| jira_advanced_setting |   ❌  |
| jira_backlog_issue    |   ✅  |
| jira_board            |   ✅  |
| jira_component        |   ❌  |
| jira_dashboard        |   ❌  |
| jira_epic             |   ❌  |
| jira_global_setting   |   ❌  |
| jira_group            |   ❌  |
| jira_issue            |   ✅  |
| jira_issue_type       |   ❌  |
| jira_priority         |   ❌  |
| jira_project          |   ❌  |
| jira_project_role     |   ❌  |
| jira_sprint           |   ✅  |
| jira_user             |   ❌  |
| jira_workflow         |   ❌  |

## Get involved

- Open source: https://github.com/turbot/steampipe-plugin-jira
- Community: [Join #steampipe on Slack →](https://turbot.com/community/join)
