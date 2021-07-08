# Table: jira_advanced_setting

Jira has a small number of commonly edited advanced configuration options, which are stored in the Jira database. These options can be accessed and edited from the Advanced Settings page. You must be a Jira System Administrator to do this.

## Examples

### Basic info

```sql
select
  id,
  name,
  key,
  type
from
  jira_advanced_setting;
```

### list advanced settings that supports string type value

```sql
select
  id,
  name,
  key,
  type
from
  jira_advanced_setting
where
  type = 'string';
```
