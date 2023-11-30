---
title: "Steampipe Table: jira_user - Query Jira Users using SQL"
description: "Allows users to query Jira Users, specifically user account details, providing insights into user activities and account statuses."
---

# Table: jira_user - Query Jira Users using SQL

Jira is a popular project management tool developed by Atlassian. It is primarily used for issue tracking, bug tracking, and project management. Jira allows teams to plan, track, and manage agile software development projects.

## Table Usage Guide

The `jira_user` table provides insights into user accounts within Jira. As a project manager or system administrator, explore user-specific details through this table, including account statuses, email addresses, and associated metadata. Utilize it to uncover information about users, such as their account statuses, their last login details, and their group memberships.

## Examples

### Basic info
Explore which Jira users are active and understand their respective account types. This can help in managing user roles and access within your Jira environment.

```sql
select
  display_name,
  account_type as type,
  active as status,
  account_id
from
  jira_user;
```

### Get associated names for a particular user
Explore the group affiliations of a specific user to understand their role and access permissions within the system.

```sql
select
  display_name,
  active as status,
  account_id,
  jsonb_pretty(group_names) as group_names
from
  jira_user
where
  display_name = 'Confluence Analytics (System)';
```