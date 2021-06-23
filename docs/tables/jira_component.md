# Table: jira_component

Jira project components are generic containers for issues. Components can have Component Leads: people who are automatically assigned issues with that component. Components add some structure to projects, breaking it up into features, teams, modules, subprojects, and more. Using components, you can generate reports, collect statistics, display it on dashboards, etc.

## Examples

### Basic info

```sql
select
  id,
  name,
  project,
  issue_count
from
  jira_component;
```

### List components having issues

```sql
select
  id,
  name,
  project,
  issue_count
from
  jira_component
where
  issue_count > 0;
```

### List components with no leads

```sql
select
  id,
  name,
  project,
  issue_count,
  lead_display_name
from
  jira_component
where
  lead_display_name = '';
```
