## v0.14.0 [2023-12-12]

_What's new?_

- The plugin can now be downloaded and used with the [Steampipe CLI](https://steampipe.io/docs), as a [Postgres FDW](https://steampipe.io/docs/steampipe_postgres/overview), as a [SQLite extension](https://steampipe.io/docs//steampipe_sqlite/overview) and as a standalone [exporter](https://steampipe.io/docs/steampipe_export/overview). ([#112](https://github.com/turbot/steampipe-plugin-jira/pull/112))
- The table docs have been updated to provide corresponding example queries for Postgres FDW and SQLite extension. ([#112](https://github.com/turbot/steampipe-plugin-jira/pull/112))
- Docs license updated to match Steampipe [CC BY-NC-ND license](https://github.com/turbot/steampipe-plugin-jira/blob/main/docs/LICENSE). ([#112](https://github.com/turbot/steampipe-plugin-jira/pull/112))

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.8.0](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v580-2023-12-11) that includes plugin server encapsulation for in-process and GRPC usage, adding Steampipe Plugin SDK version to `_ctx` column, and fixing connection and potential divide-by-zero bugs. ([#111](https://github.com/turbot/steampipe-plugin-jira/pull/111))

## v0.13.0 [2023-11-15]

_What's new?_

- New tables added
  - [jira_issue_comment](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_issue_comment) ([#103](https://github.com/turbot/steampipe-plugin-jira/pull/103))
  - [jira_issue_worklog](https://hub.steampipe.io/plugins/turbot/jira/tables/jira_issue_worklog) ([#104](https://github.com/turbot/steampipe-plugin-jira/pull/104))

_Enhancements_

- Added the `properties` column to `jira_project` table. ([#105](https://github.com/turbot/steampipe-plugin-jira/pull/105))

_Bug fixes_

- Fixed typo in the docs/index.md file. ([#102](https://github.com/turbot/steampipe-plugin-jira/pull/102)) (Thanks [@adrfrank](https://github.com/adrfrank) for the contribution!)
- Fixed the `jira_issue` table by enhancing case insensitivity support for the `status` column. ([#90](https://github.com/turbot/steampipe-plugin-jira/pull/90))

## v0.12.1 [2023-10-04]

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.6.2](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v562-2023-10-03) which prevents nil pointer reference errors for implicit hydrate configs. ([#98](https://github.com/turbot/steampipe-plugin-jira/pull/98))

## v0.12.0 [2023-10-02]

_Dependencies_

- Upgraded to [steampipe-plugin-sdk v5.6.1](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v561-2023-09-29) with support for rate limiters. ([#96](https://github.com/turbot/steampipe-plugin-jira/pull/96))
- Recompiled plugin with Go version `1.21`. ([#96](https://github.com/turbot/steampipe-plugin-jira/pull/96))

## v0.11.0 [2023-09-21]

_What's new?_

- Added support for querying on-premise Jira instances. This can be done by setting the `personal_access_token` config argument in the `jira.spc` file. ([#86](https://github.com/turbot/steampipe-plugin-jira/pull/86)) (Thanks [@juandspy](https://github.com/juandspy) for the contribution!)

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.5.1](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v551-2023-07-26) which significantly reduces API calls and boosts query performance, resulting in faster data retrieval. ([#92](https://github.com/turbot/steampipe-plugin-jira/pull/92))

## v0.10.2 [2023-08-22]

_Bug fixes_

- Fixed the `sprint_ids` column of the `jira_issue` table to correctly return data for issues that have no sprints, instead of returning an error. ([#87](https://github.com/turbot/steampipe-plugin-jira/pull/87)) (Thanks [@juandspy](https://github.com/juandspy) for the contribution!)

## v0.10.1 [2023-06-16]

_Bug fixes_

- Fixed pagination in the `jira_issue` table to avoid the truncation of the result set. ([#82](https://github.com/turbot/steampipe-plugin-jira/pull/82))

## v0.10.0 [2023-06-14]

_Enhancements_

- Added `JIRA_URL`, `JIRA_USER` and `JIRA_TOKEN` environment variables for setting `base_url`, `username` and `token` config arguments respectively. ([#79](https://github.com/turbot/steampipe-plugin-jira/pull/79))

_Bug fixes_

- Fixed the `epic_key` column in `jira_issue` table to consistently return data instead of `null` when key columns are not passed in the `where` clause. ([#80](https://github.com/turbot/steampipe-plugin-jira/pull/80))

## v0.9.0 [2023-04-10]

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.3.0](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v530-2023-03-16) which includes fixes for query cache pending item mechanism and aggregator connections not working for dynamic tables. ([#77](https://github.com/turbot/steampipe-plugin-jira/pull/77))

## v0.8.0 [2022-11-24]

_Enhancements_

- Added support for `DefaultRetryConfig` configuration across all the tables to retry `429 Too Many Requests` (Rate Limiting) errors. ([#72](https://github.com/turbot/steampipe-plugin-jira/pull/72))

## v0.7.0 [2022-11-17]

_Enhancements_

- Added the `status_category` column to the `jira_issue` table. ([#69](https://github.com/turbot/steampipe-plugin-jira/pull/69)) (Thanks to [@gabrielsoltz](https://github.com/gabrielsoltz) for the contribution!)

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v4.1.8](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v418-2022-09-08) which increases the default open file limit. ([#70](https://github.com/turbot/steampipe-plugin-jira/pull/70))

## v0.6.1 [2022-10-13]

_Bug fixes_

- Not found errors are now handled correctly in `jira_backlog_issue`, `jira_component`, `jira_issue_type`, and `jira_sprint` tables.

## v0.6.0 [2022-09-26]

_Bug fixes_

- Fixed typos in the `plugin.go` file and updated the filename to use `jira/table_jira_global_setting.go` instead of `jira/table_ jira_global_setting.go`. ([#58](https://github.com/turbot/steampipe-plugin-jira/pull/58)) (Thanks to [@s-spindler](https://github.com/s-spindler) for the contribution!)

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v4.1.7](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v417-2022-09-08) which includes several caching and memory management improvements. ([#62](https://github.com/turbot/steampipe-plugin-jira/pull/62))
- Recompiled plugin with Go version `1.19`. ([#62](https://github.com/turbot/steampipe-plugin-jira/pull/62))

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
