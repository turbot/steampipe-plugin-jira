---
title: "Steampipe Table: jira_advanced_setting - Query Jira Advanced Settings using SQL"
description: "Allows users to query Advanced Settings in Jira, specifically the key and value pairs of settings, providing insights into system configurations and potential modifications."
---

# Table: jira_advanced_setting - Query Jira Advanced Settings using SQL

Jira Advanced Settings is a feature within Atlassian's Jira that allows you to customize and configure the system to suit your organization's needs. It provides a way to set up and manage key-value pairs for various settings, including index path, attachment size, and more. Jira Advanced Settings helps you stay informed about the current configurations and take appropriate actions when modifications are needed.

## Table Usage Guide

The `jira_advanced_setting` table provides insights into the advanced settings within Jira. As a system administrator, explore setting-specific details through this table, including the key and value pairs of each setting. Utilize it to uncover information about system configurations, such as attachment size, index path, and the possibility of potential modifications.

## Examples

### Basic info
Explore the advanced settings in your Jira instance to understand their types and associated keys. This can be useful in assessing the current configuration and identifying areas for optimization or troubleshooting.

```sql
select
  id,
  name,
  key,
  type
from
  jira_advanced_setting;
```

### list advanced settings that supports string type value
Explore advanced settings within Jira that support string type values. This can be useful for configuring and customizing your Jira environment to best suit your needs.

```sql
select
  id,
  name,
  key,
  type
from
  jira_advanced_setting
where
  type = 'string';
```