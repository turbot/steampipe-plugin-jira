---
title: "Steampipe Table: jira_dashboard - Query Jira Dashboards using SQL"
description: "Allows users to query Jira Dashboards, providing insights into the various dashboards available in a Jira Software instance."
---

# Table: jira_dashboard - Query Jira Dashboards using SQL

Jira Software is a project management tool developed by Atlassian. It provides a platform for planning, tracking, and releasing software, and is widely used by agile teams. A key feature of Jira Software are dashboards, which provide a customizable and flexible view of a project's progress and status.

## Table Usage Guide

The `jira_dashboard` table provides insights into the various dashboards available within a Jira Software instance. As a project manager or team lead, explore dashboard-specific details through this table, including the owner, viewability, and associated projects. Utilize it to uncover information about dashboards, such as those that are shared with everyone, the ones owned by a specific user, and the projects associated with each dashboard.

## Examples

### Basic info
Gain insights into your favorite Jira dashboards and their respective owners. This can help you understand who is responsible for the dashboards you frequently use.

```sql+postgres
select
  id,
  name,
  is_favourite,
  owner_account_id,
  owner_display_name
from
  jira_dashboard;
```

```sql+sqlite
select
  id,
  name,
  is_favourite,
  owner_account_id,
  owner_display_name
from
  jira_dashboard;
```

### Get share permissions details
Explore which Jira dashboards have specific share permissions. This can help you understand how information is being disseminated, ensuring the right teams have access to the right data.

```sql+postgres
select
  id,
  name,
  owner_display_name,
  popularity,
  jsonb_pretty(share_permissions) as share_permissions
from
  jira_dashboard;
```

```sql+sqlite
select
  id,
  name,
  owner_display_name,
  popularity,
  share_permissions
from
  jira_dashboard;
```