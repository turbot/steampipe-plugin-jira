# Table: jira_priority

An issue's priority defines its importance in relation to other issues, so it helps users to determine which issues should be tackled first.
Jira comes with a set of default priorities, which you can modify or add to. You can also choose different priorities for your projects.

## Examples

### Basic info

```sql
select
  name,
  id,
  description
from
  jira_priority;
```

### List issues with high priority

```sql
select
  id as issue_no,
  description as issue_description,
  assignee_display_name as assigned_to
from
  jira_issue
where 
  priority = 'High';
```

### Count of issues per priority

```sql
select
  p.name as priority,
  count(i.id) as issue_count
from
  jira_priority as p
  left join jira_issue as i on i.priority = p.name
group by p.name;
```
