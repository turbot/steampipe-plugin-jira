## v0.5.0 [2022-06-17]

_Enhancements_

- Added the column `resolution_date` to `jira_issue` table. ([#56](https://github.com/turbot/steampipe-plugin-jira/pull/56))

## v0.4.1 [2022-05-23]

_Bug fixes_

- Fixed the Slack community links in README and docs/index.md files. ([#53](https://github.com/turbot/steampipe-plugin-jira/pull/53))

## v0.4.0 [2022-04-27]

_Enhancements_

- Added support for native Linux ARM and Mac M1 builds. ([#51](https://github.com/turbot/steampipe-plugin-jira/pull/51))

## v0.3.0 [2022-04-25]

_Enhancements_

- Added `key` and `project_type_key` as optional list key columns to `jira_project` table. ([#50](https://github.com/turbot/steampipe-plugin-jira/pull/50))
- Added context cancellation and limit handling for all tables. ([#50](https://github.com/turbot/steampipe-plugin-jira/pull/50))
- Improved help messages if any of the require configuration arguments aren't set. ([#50](https://github.com/turbot/steampipe-plugin-jira/pull/50))

_Bug fixes_

- Fixed paging for `jira_project` and `jira_user` tables so all results should be returned correctly. ([#50](https://github.com/turbot/steampipe-plugin-jira/pull/50))

## v0.2.0 [2022-04-15]

_Enhancements_

- Added optional quals for `jira_issue` table ([#47](https://github.com/turbot/steampipe-plugin-jira/pull/47))
- Recompiled plugin with [steampipe-plugin-sdk v3.1.0](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v310--2022-03-30) ([#47](https://github.com/turbot/steampipe-plugin-jira/pull/47))

## v0.1.0 [2021-11-23]

_Enhancements_

- Recompiled plugin with Go version 1.17 ([#44](https://github.com/turbot/steampipe-plugin-jira/pull/44))
- Recompiled plugin with [steampipe-plugin-sdk v1.8.2](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v182--2021-11-22) ([#43](https://github.com/turbot/steampipe-plugin-jira/pull/43))

## v0.0.3 [2021-09-22]

_What's new?_

- New tables added
  - [jira_global_setting](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_global_setting) ([#23](https://github.com/turbot/steampipe-plugin-jira/pull/23))
  - [jira_workflow](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_workflow) ([#30](https://github.com/turbot/steampipe-plugin-jira/pull/30))

_Enhancements_

- Recompiled plugin with [steampipe-plugin-sdk v1.6.1](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v161--2021-09-21) ([#40](https://github.com/turbot/steampipe-plugin-jira/pull/40))

## v0.0.2 [2021-07-08]

_What's new?_

- New tables added
  - [jira_advanced_setting](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_advanced_setting) ([#18](https://github.com/turbot/steampipe-plugin-jira/pull/18))
  - [jira_backlog_issue](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_backlog_issue) ([#16](https://github.com/turbot/steampipe-plugin-jira/pull/16))
  - [jira_component](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_component) ([#15](https://github.com/turbot/steampipe-plugin-jira/pull/15))
  - [jira_issue_type](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_issue_type) ([#17](https://github.com/turbot/steampipe-plugin-jira/pull/17))
  - [jira_priority](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_priority) ([#14](https://github.com/turbot/steampipe-plugin-jira/pull/14))

_Enhancements_

- Updated: Plugin category is now `software development`

_Bug fixes_

- Fixed: Cleanup unused code ([#21](https://github.com/turbot/steampipe-plugin-jira/pull/21))

## v0.0.1 [2021-06-17]

_What's new?_

- New tables added

  - [jira_board](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_board)
  - [jira_dashboard](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_dashboard)
  - [jira_epic](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_epic)
  - [jira_group](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_group)
  - [jira_issue](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_issue)
  - [jira_project](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_project)
  - [jira_project_role](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_project_role)
  - [jira_sprint](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_sprint)
  - [jira_user](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_user)
