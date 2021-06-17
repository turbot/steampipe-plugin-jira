# Table: jira_project

A project is simply a collection of issues (stories, bugs, tasks, etc). You would typically use a project to represent the development work for a product, project, or service in Jira Software.

## Examples

### Basic info

```sql
select
  name,
  id,
  key,
  lead_display_name,
  project_category,
  description
from
  jira_project;
```

### List all issues in a project

```sql
select
  id,
  key,
  project_id,
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
