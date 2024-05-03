---
title: "Steampipe Table: jira_issue - Query Jira Issues using SQL"
description: "Allows users to query Jira Issues, specifically providing insights into issue details such as status, assignee, reporter, project, and more."
---

# Table: jira_issue - Query Jira Issues using SQL

Jira is a project management tool developed by Atlassian, widely used for issue tracking, bug tracking, and agile project management. It allows teams to manage, plan, track, and release software, ensuring transparency and team collaboration. Jira's issues, the core units of Jira, are used to track individual pieces of work that need to be completed.

## Table Usage Guide

The `jira_issue` table provides insights into Jira issues within a project. As a project manager or software developer, explore issue-specific details through this table, including status, assignee, reporter, and associated metadata. Utilize it to uncover information about issues, such as those unassigned, those in progress, and to verify project timelines.

## Examples

### Basic info
Discover the segments that detail the creation and status of a project. This is useful to gain insights into project timelines, creators, and their current progress.

```sql+postgres
select
  key,
  project_key,
  created,
  creator_display_name,
  status,
  summary
from
  jira_issue;
```

```sql+sqlite
select
  key,
  project_key,
  created,
  creator_display_name,
  status,
  summary
from
  jira_issue;
```

### List issues for a specific project
Explore the status and details of issues related to a specific project. This can aid in understanding the project's progress, identifying any roadblocks, and planning further actions effectively.

```sql+postgres
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
  jira_issue
where
  project_key = 'TEST';
```

```sql+sqlite
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
  jira_issue
where
  project_key = 'TEST';
```

### List all issues assigned to a specific user
Explore which tasks are allocated to a particular individual, allowing you to gain insights into their workload and responsibilities. This is particularly useful for project management and task distribution.

```sql+postgres
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id
from
  jira_issue
where
  assignee_display_name = 'Lalit Bhardwaj';
```

```sql+sqlite
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id
from
  jira_issue
where
  assignee_display_name = 'Lalit Bhardwaj';
```

### List issues due in the next week
Explore upcoming tasks by identifying issues scheduled for completion within the next week. This can help in prioritizing work and managing team assignments effectively.
```sql+postgres
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id,
  duedate
from
  jira_issue
where
  duedate > current_date
  and duedate <= (current_date + interval '7' day);
```

```sql+sqlite
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id,
  duedate
from
  jira_issue
where
  duedate > date('now')
  and duedate <= date('now', '+7 day');
```

#### Get linked issue details for the issues
The query enables deep analysis of linked issues within Jira. Each issue can have multiple links to other issues, which might represent dependencies, blockers, duplicates, or other relationships critical for project management and bug tracking.
```sql+postgres
select
  ji.id,
  ji.title,
  ji.project_key,
  ji.status,
  il.issue_link_id,
  il.issue_link_self,
  il.issue_link_type_name,
  il.inward_issue_id,
  il.inward_issue_key,
  il.inward_issue_status_name,
  il.inward_issue_summary,
  il.inward_issue_priority_name
from
  jira_issue ji,
  lateral jsonb_array_elements(ji.fields -> 'issuelinks') as il_data,
  lateral (
    select
      il_data ->> 'id' as issue_link_id,
      il_data ->> 'self' as issue_link_self,
      il_data -> 'type' ->> 'name' as issue_link_type_name,
      il_data -> 'inwardIssue' ->> 'id' as inward_issue_id,
      il_data -> 'inwardIssue' ->> 'key' as inward_issue_key,
      il_data -> 'inwardIssue' -> 'fields' -> 'status' ->> 'name' as inward_issue_status_name,
      il_data -> 'inwardIssue' -> 'fields' ->> 'summary' as inward_issue_summary,
      il_data -> 'inwardIssue' -> 'fields' -> 'priority' ->> 'name' as inward_issue_priority_name
  ) as il;
```

```sql+sqlite
select
  ji.id,
  ji.title,
  ji.project_key,
  ji.status,
  json_extract(il_data.value, '$.id') as issue_link_id,
  json_extract(il_data.value, '$.self') as issue_link_self,
  json_extract(il_data.value, '$.type.name') as issue_link_type_name,
  json_extract(il_data.value, '$.inwardIssue.id') as inward_issue_id,
  json_extract(il_data.value, '$.inwardIssue.key') as inward_issue_key,
  json_extract(il_data.value, '$.inwardIssue.fields.status.name') as inward_issue_status_name,
  json_extract(il_data.value, '$.inwardIssue.fields.summary') as inward_issue_summary,
  json_extract(il_data.value, '$.inwardIssue.fields.priority.name') as inward_issue_priority_name
from
 jira_issue ji,
 json_each(json_extract(ji.fields, '$.issuelinks')) as il_data;
```



### Get issues belonging to a sprint
1. "Explore which tasks are part of a particular sprint, allowing you to manage and prioritize your team's workflow effectively."
2. "Identify all tasks that have been completed, providing a clear overview of your team's accomplishments and productivity."
3. "Determine the tasks that are currently awaiting support, helping you to allocate resources and address bottlenecks in your workflow."
4. "Review the different status categories within a specific project, offering insights into the project's progress and potential areas for improvement.

```sql+postgres
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id,
  duedate
from
  jira_issue
where
  sprint_ids @> '2';
```

```sql+sqlite
Error: SQLite does not support array contains operator '@>'.
```

#### List all issues in status category 'Done'

```sql+postgres
select
  id,
  key,
  summary,
  status,
  status_category,
  assignee_display_name
from
  jira_issue
where
  status_category = 'Done';
```

```sql+sqlite
select
  id,
  key,
  summary,
  status,
  status_category,
  assignee_display_name
from
  jira_issue
where
  status_category = 'Done';
```

#### List all issues in status Waiting for Support

```sql+postgres
select
  id,
  key,
  summary,
  status,
  status_category,
  assignee_display_name
from
  jira_issue
where
  status = 'Waiting for support';
```

```sql+sqlite
select
  id,
  key,
  summary,
  status,
  status_category,
  assignee_display_name
from
  jira_issue
where
  status = 'Waiting for support';
```

#### List all possible status for each status_category for a speficic project

```sql+postgres
select distinct
  project_key,
  status_category,
  status
from
  jira_issue
where
  project_key = 'PROJECT-KEY'
order by
  status_category;
```

```sql+sqlite
select distinct
  project_key,
  status_category,
  status
from
  jira_issue
where
  project_key = 'PROJECT-KEY'
order by
  status_category;
```
