# Table: jira_epic

An **Epic** is essentially a large user story that can be broken down into a number of smaller stories. An epic can span more than one project.

## Examples

### Basic info

```sql
select
  id,
  name,
  key,
  done as status,
  summary
from
  jira_epic;
```

### List issues in epic

```sql
select
  i.id as issue_id,
  i.status as issue_status,
  i.created as issue_created,
  i.creator_display_name,
  i.assignee_display_name,
  e.id as epic_id,
  e.name as epic_name,
  i.summary as issue_summary
from
  jira_epic as e,
  jira_issue as i
where
  i.epic_key = e.key;
```
