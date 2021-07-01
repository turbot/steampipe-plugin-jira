# Table: jira_backlog_issue

**Issues** are the building blocks of any Jira project. The backlog contains incomplete issues that are not assigned to any future or active sprint.

## Examples

### Basic info

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
