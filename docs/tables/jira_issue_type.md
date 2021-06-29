# Table: jira_issue_type

**Issue types** distinguish different types of work in unique ways, and helps to identify, categorize, and report on the team’s work across the Jira site.

## Examples

### Basic info

```sql
select
  id,
  name,
  description,
  avatar_id,
from
  jira_issue_type;
```

### List issue types for a specific project

```sql
select
  id,
  name,
  description,
  avatar_id,
  scope
from
  jira_issue_type
where
  scope -> 'project' ->> 'id' = '10000';
```

### List issue types associated with subtask creation
```sql
select
  id,
  name,
  description,
  avatar_id,
  subtask
from
  jira_issue_type
where
  subtask;
```

### List issue types with hierarchy_level 0 (Base)

```sql
select
  id,
  name,
  description,
  avatar_id,
  hierarchy_level
from
  jira_issue_type
where
  hierarchy_level = '0';
```
