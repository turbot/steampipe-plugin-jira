# Table: jira_project_role

Project roles are a flexible way to associate users and/or groups with particular projects. Project roles also allow for delegated administration:

- Jira administrators define project roles — that is, all projects have the same project roles available to them.
- Project administrators assign members to project roles specifically for their project(s).

  A project administrator is someone who has the project-specific 'Administer Project' permission, but not necessarily the global 'Jira Administrator' permission.

**Note:** Project roles are somewhat similar to groups, the main difference being that group membership is global whereas project role membership is project-specific. Additionally, group membership can only be altered by Jira administrators, whereas project role membership can be altered by project administrators.

## Examples

### Basic info

```sql
select
  id,
  name,
  description
from
  jira_project_role;
```

### Get actor details

```sql
select
  id,
  name,
  jsonb_pretty(actor_account_ids) as actor_account_ids,
  jsonb_pretty(actor_names) as actor_names
from
  jira_project_role;
```

### Get actor details joined with user table

```sql
select
  id,
  name,
  actor_id,
  actor.display_name,
  actor.account_type,
  actor.active as actor_status
from
  jira_project_role as role,
  jsonb_array_elements_text(actor_account_ids) as actor_id,
  jira_user as actor
where
  actor_id = actor.account_id;
```
