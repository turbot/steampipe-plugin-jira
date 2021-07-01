# Table: jira_global_setting

Jira global settings contains options to check and configure settings in your Jira installation that are applied globally to all users.

## Example

### Basic info

```sql
select
  voting_enabled,
  watching_enabled,
  sub_tasks_enabled,
  time_tracking_enabled,
  time_tracking_configuration
from
  jira_global_setting;
```
