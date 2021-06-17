# Table: jira_user

## Examples

### Basic info

```sql
select
  display_name,
  account_type as type,
  active as status,
  account_id
from
  jira_user;
```

### Get associated names for a particular user

```sql
select
  display_name,
  active as status,
  account_id,
  jsonb_pretty(group_names) as group_names
from
  jira_user
where
  display_name = 'Confluence Analytics (System)';
```
