---
title: "Steampipe Table: jira_backlog_issue - Query Jira Backlog Issues using SQL"
description: "Allows users to query Jira Backlog Issues, providing an overview of all issues currently in the backlog of a Jira project."
---

# Table: jira_backlog_issue - Query Jira Backlog Issues using SQL

Jira is a project management tool used for issue tracking, bug tracking, and agile project management. A Jira Backlog Issue refers to a task or a bug that has been identified but is not currently being worked on. These issues are stored in the backlog, a list of tasks or bugs that need to be addressed but have not yet been prioritized for action.

## Table Usage Guide

The `jira_backlog_issue` table provides insights into the backlog issues within a Jira project. As a project manager or a software developer, you can use this table to explore details of each issue, including its status, priority, and assignee. This can help you prioritize tasks, manage project workflows, and ensure timely resolution of bugs and tasks.

## Examples

### Basic info
Explore which projects have been created, who initiated them, their current status, and a brief summary. This information can be useful to gain an overview of ongoing projects and their progress.

```sql
select
  key,
  project_key,
  created,
  creator_display_name,
  status,
  summary
from
  jira_backlog_issue;
```

### List backlog issues for a specific project
Explore the status and details of pending tasks within a specific project to manage workload and track progress effectively. This can help in prioritizing tasks and assigning them to the right team members.

```sql
select
  id,
  key,
  project_key,
  created,
  creator_display_name,
  assignee_display_name,
  status,
  summary
from
  jira_backlog_issue
where
  project_key = 'TEST1';
```

### List backlog issues assigned to a specific user
Explore which backlog issues are assigned to a specific user to manage and prioritize their workload efficiently. This is useful in tracking project progress and ensuring tasks are evenly distributed among team members.

```sql
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id
from
  jira_backlog_issue
where
  assignee_display_name = 'sayan';
```

### List backlog issues due in 30 days
Explore which backlog issues are due within the next 30 days to better manage your project timeline and delegate tasks effectively. This can assist in prioritizing work and ensuring that deadlines are met.

```sql
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id,
  due_date
from
  jira_backlog_issue
where
  due_date > current_date
  and due_date <= (current_date + interval '30' day);
```