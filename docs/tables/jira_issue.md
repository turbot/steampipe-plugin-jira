# Table:

Issues are the building blocks of any Jira project. An issue could represent a story, a bug, a task, or another issue type in your project.

## Examples

### List issues for a specific project(can use column project_key or project_id )

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
  jira_issue
where
  project_key = 'TEST';
```

### List all issues assignment to a user

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
  jira_issue
where
  assignee_display_name = 'Lalit Bhardwaj';
```
