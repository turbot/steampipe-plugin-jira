---
title: "Steampipe Table: jira_priority - Query Jira Priorities using SQL"
description: "Allows users to query Jira Priorities, specifically providing details about the priority levels assigned to different issues in a Jira project."
---

# Table: jira_priority - Query Jira Priorities using SQL

Jira is a project management tool used for issue tracking, bug tracking, and agile project management. The priority of an issue in Jira signifies its importance in relation to other issues. It is an attribute that can be set by the user when creating or editing issues.

## Table Usage Guide

The `jira_priority` table provides insights into the priority levels assigned to different issues within a Jira project. As a project manager or agile team member, explore priority-specific details through this table, including descriptions, icons, and associated metadata. Utilize it to uncover information about priorities, such as their relative importance, to help in effective issue management and resolution.

## Examples

### Basic info
Explore the priorities set in your Jira project management tool. This can help you understand how tasks are being prioritized, assisting in better project management and resource allocation.

```sql+postgres
select
  name,
  id,
  description
from
  jira_priority;
```

```sql+sqlite
select
  name,
  id,
  description
from
  jira_priority;
```

### List issues with high priority
Discover the segments that have been assigned high priority issues in order to prioritize your team's workflow and address critical tasks more efficiently.

```sql+postgres
select
  id as issue_no,
  description as issue_description,
  assignee_display_name as assigned_to
from
  jira_issue
where 
  priority = 'High';
```

```sql+sqlite
select
  id as issue_no,
  description as issue_description,
  assignee_display_name as assigned_to
from
  jira_issue
where 
  priority = 'High';
```

### Count of issues per priority
Determine the distribution of issues across different priority levels to better understand where the majority of concerns lie. This can help in prioritizing resources and efforts for issue resolution.

```sql+postgres
select
  p.name as priority,
  count(i.id) as issue_count
from
  jira_priority as p
  left join jira_issue as i on i.priority = p.name
group by p.name;
```

```sql+sqlite
select
  p.name as priority,
  count(i.id) as issue_count
from
  jira_priority as p
  left join jira_issue as i on i.priority = p.name
group by p.name;
```