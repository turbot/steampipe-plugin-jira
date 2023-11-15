# Table: jira_issue_worklog

The Jira issue worklog is a feature in Atlassian's Jira software that allows users to record and track the amount of time they have spent working on various tasks or issues. This is particularly useful in project management and software development contexts, where tracking time spent on tasks is crucial for understanding project progress, billing, and workload distribution.

## Examples

### Basic info

```sql
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

```sql
select
  issue_id,
  sum(time_spent_seconds) as total_time_spent_seconds
from
  jira_issue_worklog
group by
  issue_id;
```

### Show the latest worklogs for issues from the past 5 days

```sql
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

### Retrieve issues and their worklogs updated in the last 10 days

```sql
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

### Get author information of worklogs

```sql
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
