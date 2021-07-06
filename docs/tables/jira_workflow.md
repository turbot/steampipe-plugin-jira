# Table: jira_workflow

A **Workflow** is a set of statuses and transitions that an issue moves through during its lifecycle, and typically represents a process within your organization. Workflows can be associated with particular projects and, optionally, specific issue types by using a workflow scheme.
## Examples

### Basic info

```sql
select
  name,
  entity_id,
  description,
  is_default
from
  jira_workflow;
```

### List workflows that are not default

```sql
select
  name,
  entity_id,
  description,
  is_default
from
  jira_workflow
where
  not is_default;
```

### List workflows that are not associated with entity id

```sql
select
  name,
  entity_id,
  description,
  is_default
from
  jira_workflow
where
  entity_id = '';
```
