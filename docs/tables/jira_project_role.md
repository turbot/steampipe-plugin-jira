---
title: "Steampipe Table: jira_project_role - Query Jira Project Roles using SQL"
description: "Allows users to query Jira Project Roles, providing insights into role details, permissions, and associated users and groups."
---

# Table: jira_project_role - Query Jira Project Roles using SQL

Jira Project Roles are a flexible way to associate users and groups with projects. They allow project administrators to manage project role membership. Project roles can be used in permission schemes, issue security levels, notification schemes, and comment visibility.

## Table Usage Guide

The `jira_project_role` table provides insights into project roles within Jira. As a project administrator, explore role-specific details through this table, including permissions and associated users and groups. Utilize it to manage project role membership, and to set up permission schemes, issue security levels, and notification schemes.

**Important Notes**
- Project roles are somewhat similar to groups, the main difference being that group membership is global whereas project role membership is project-specific. Additionally, group membership can only be altered by Jira administrators, whereas project role membership can be altered by project administrators.

## Examples

### Basic info
Explore the different roles within your Jira project. This can help in understanding the distribution of responsibilities and in managing team members more effectively.

```sql+postgres
select
  id,
  name,
  description
from
  jira_project_role;
```

```sql+sqlite
select
  id,
  name,
  description
from
  jira_project_role;
```

### Get actor details
Explore the details of different actors within your Jira project roles. This query is useful for gaining insights into the identities and account IDs of actors, aiding in project management and team coordination.

```sql+postgres
select
  id,
  name,
  jsonb_pretty(actor_account_ids) as actor_account_ids,
  jsonb_pretty(actor_names) as actor_names
from
  jira_project_role;
```

```sql+sqlite
select
  id,
  name,
  actor_account_ids,
  actor_names
from
  jira_project_role;
```

### Get actor details joined with user table
This query is used to identify the details of actors from the user table in a Jira project. It can be useful in understanding the roles and statuses of different actors in the project, which can aid project management and team coordination.

```sql+postgres
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

```sql+sqlite
select
  role.id,
  role.name,
  actor_id.value as actor_id,
  actor.display_name,
  actor.account_type,
  actor.active as actor_status
from
  jira_project_role as role,
  json_each(role.actor_account_ids) as actor_id,
  jira_user as actor
where
  actor_id.value = actor.account_id;
```