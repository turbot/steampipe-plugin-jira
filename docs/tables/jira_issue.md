# Table:

Issues are the building blocks of any Jira project. An issue could represent a story, a bug, a task, or another issue type in your project.

This tables requires an '=' qualifier for the following column: project_key

## Examples

### List issues for a specific project

```sql
select
  id,
  key,
  project_key,
  created,
  created,
  assignee,
  status,
  summary
from
  jira_issue
where
  project_key = 'TEST';
```

### List all issues

```sql
select
  i.id,
  i.key,
  i.project_key,
  i.created,
  i.created,
  i.assignee,
  i.status,
  i.summary
from
  jira_project p,
  jira_issue i
where
  i.project_key = p.key
order by
  id;
```
