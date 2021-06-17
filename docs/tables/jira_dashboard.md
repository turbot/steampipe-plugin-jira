# Table:

Dashboard is the main display you see when you log in to Jira. You can create
multiple dashboards from different projects, or multiple dashboards for one
massive overview of all the work you're involved with.

Dashboards are designed to display gadgets that help you organize your
projects, assignments, and achievements in different charts.

## Examples

### Basic info

```sql
select
  id,
  name,
  is_favourite,
  owner_account_id,
  owner_display_name
from
  jira_dashboard;
```

### Get share permissions details

```sql
select
  id,
  name,
  owner_display_name,
  popularity,
  jsonb_pretty(share_permissions) as share_permissions
from
  jira_dashboard;
```
