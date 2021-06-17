# Table: jira_sprint

A **Sprint** — also known as an iteration — is a short period in which the development team implements and delivers a discrete and potentially shippable application increment, e.g. a working milestone version.

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

### List sprints due in the next week

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
  end_date > current_date
  and end_date <= (current_date + interval '7' day);
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
  sprint_ids @> '2';
```

### Count of issues by sprint
```sql
select 
  b.name as board_name,
  s.name as sprint_name,
  count(sprint_id) as num_issues
from 
  jira_board as b
  join jira_sprint as s on s.board_id = b.id 
  left join jira_issue as i on true 
  left join jsonb_array_elements(i.sprint_ids) as sprint_id on sprint_id ::bigint = s.id  
group by
  board_name,
  sprint_name
order by
  board_name,
  sprint_name;
```