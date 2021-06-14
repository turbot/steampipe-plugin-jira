# Table: jira_group

Group is a collection of users. Administrators create groups so that the administrator can assign permissions to a number of people at once.

## Examples

### Basic info

```sql
select
  name,
  id
from
  jira_group;
```

### Get associated users

```sql
select
  name as group_name,
  u.display_name as user_name,
  user_id,
  u.email_address as user_email
from
  jira_group g,
  jsonb_array_elements_text(member_ids) as user_id
  join jira_user u on user_id = u.account_id
order by name, user_name;
```
