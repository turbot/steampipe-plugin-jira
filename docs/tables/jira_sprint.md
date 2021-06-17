# Table: jira_sprint

A sprint — also known as an iteration — is a short period in which the development team implements and delivers a discrete and potentially shippable application increment, e.g. a working milestone version.

## Examples

### Basic info

```sql
select
  id,
  name,
  board_id,
  state,
  start_date,
  end_date,
  complete_date
from
  jira_sprint;
```

### List sprints due in next week

```sql
select
  id,
  name,
  board_id,
  state,
  start_date,
  end_date
from
  jira_sprint
where
  end_date > (current_date + interval '7' day)
  and end_date <= (current_date + interval '14' day);
```

### List active sprints

```sql
select
  id,
  name,
  board_id,
  start_date,
  end_date
from
  jira_sprint
where
  state = 'active';
```

### List issues in a sprints

```sql
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
  sprint_ids @> '2'
```
