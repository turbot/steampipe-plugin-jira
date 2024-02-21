---
title: "Steampipe Table: jira_issue_worklog - Query Jira Issue Worklogs using SQL"
description: "Allows users to query Jira Issue Worklogs, specifically the work done and the time spent on each issue, providing insights into project progress and individual contributions."
---

# Table: jira_issue_worklog - Query Jira Issue Worklogs using SQL

Jira is a project management tool that enables teams to plan, track, and manage their projects and tasks. Issue Worklogs in Jira represent the work done and the time spent on each issue. This information is crucial for tracking project progress, understanding individual contributions, and managing overall project timelines.

## Table Usage Guide

The `jira_issue_worklog` table provides insights into the work done and the time spent on each issue in Jira. As a project manager or team leader, explore issue-specific details through this table, including the time spent, the work done, and the associated metadata. Utilize it to track project progress, understand individual contributions, and manage project timelines effectively.

## Examples

### Basic info
Explore the work logs of different issues in Jira to understand who has made contributions and the nature of their input. This could be useful for tracking team progress, identifying bottlenecks, and understanding individual work patterns.

```sql+postgres
select
  id,
  self,
  issue_id,
  comment,
  author
from
  jira_issue_worklog;
```

```sql+sqlite
select
  id,
  self,
  issue_id,
  comment,
  author
from
  jira_issue_worklog;
```

### Get time logged for issues
Explore which issues have had the most time logged on them. This can help prioritize resources by identifying which issues are consuming more time.

```sql+postgres
select
  issue_id,
  sum(time_spent_seconds) as total_time_spent_seconds
from
  jira_issue_worklog
group by
  issue_id;
```

```sql+sqlite
select
  issue_id,
  sum(time_spent_seconds) as total_time_spent_seconds
from
  jira_issue_worklog
group by
  issue_id;
```

### Show the latest worklogs for issues from the past 5 days
Explore the recent workload by identifying issues that have been active in the past 5 days. This can help in assessing the work distribution and identifying potential bottlenecks in real-time.

```sql+postgres
select
  id,
  issue_id,
  time_spent,
  created
from
  jira_issue_worklog
where
  created >= now() - interval '5' day;
```

```sql+sqlite
select
  id,
  issue_id,
  time_spent,
  created
from
  jira_issue_worklog
where
  created >= datetime('now', '-5 days');
```

### Retrieve issues and their worklogs updated in the last 10 days
Analyze the settings to understand the recent workload changes and their impact on project priorities. This query is useful for tracking the progress of projects and tasks that have been updated or modified in the last 10 days.

```sql+postgres
select distinct
  w.issue_id,
  w.id,
  w.time_spent,
  w.updated as worklog_updated_at,
  i.duedate,
  i.priority,
  i.project_name,
  i.key
from
  jira_issue_worklog as w,
  jira_issue as i
where
  i.id like trim(w.issue_id)
and
  w.updated >= now() - interval '10' day;
```

```sql+sqlite
select distinct
  w.issue_id,
  w.id,
  w.time_spent,
  w.updated as worklog_updated_at,
  i.duedate,
  i.priority,
  i.project_name,
  i.key
from
  jira_issue_worklog as w,
  jira_issue as i
where
  i.id like trim(w.issue_id)
and
  w.updated >= datetime('now', '-10 days');
```

### Get author information of worklogs
Discover the segments that allow you to gain insights into the authors of worklogs, such as their account type, name, email address, and time zone. This is particularly useful for understanding who is contributing to specific issues and their geographical location.

```sql+postgres
select
  id,
  issue_id,
  author ->> 'accountId' as author_account_id,
  author ->> 'accountType' as author_account_type,
  author ->> 'displayName' as author_name,
  author ->> 'emailAddress' as author_email_address,
  author ->> 'timeZone' as author_time_zone
from
  jira_issue_worklog;
```

```sql+sqlite
select
  id,
  issue_id,
  json_extract(author, '$.accountId') as author_account_id,
  json_extract(author, '$.accountType') as author_account_type,
  json_extract(author, '$.displayName') as author_name,
  json_extract(author, '$.emailAddress') as author_email_address,
  json_extract(author, '$.timeZone') as author_time_zone
from
  jira_issue_worklog;
```