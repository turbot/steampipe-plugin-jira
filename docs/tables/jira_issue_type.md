---
title: "Steampipe Table: jira_issue_type - Query Jira Issue Types using SQL"
description: "Allows users to query Jira Issue Types, providing detailed information about issue types in a Jira project, including their name, description, and avatar URL."
---

# Table: jira_issue_type - Query Jira Issue Types using SQL

Jira Issue Types are a way to categorize different types of work items in a Jira project. They help in distinguishing different types of tasks, bugs, stories, epics, and more, enabling teams to organize, track, and manage their work efficiently. Each issue type can be customized to suit the specific needs of the project or team.

## Table Usage Guide

The `jira_issue_type` table provides insights into Jira Issue Types within a project. As a project manager or a team lead, explore issue type details through this table, including their descriptions, names, and avatar URLs. Utilize it to get a comprehensive view of the different issue types in your project, aiding in better project management and task organization.

## Examples

### Basic info
Explore the different types of issues in your Jira project. This helps you to understand the variety of tasks or problems that your team handles, providing clarity on the project's complexity and scope.

```sql+postgres
select
  id,
  name,
  description,
  avatar_id
from
  jira_issue_type;
```

```sql+sqlite
select
  id,
  name,
  description,
  avatar_id
from
  jira_issue_type;
```

### List issue types for a specific project
Determine the types of issues associated with a specific project. This allows for a better understanding of the project's scope and potential challenges.

```sql+postgres
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

```sql+sqlite
select
  id,
  name,
  description,
  avatar_id,
  scope
from
  jira_issue_type
where
  json_extract(json_extract(scope, '$.project'), '$.id') = '10000';
```

### List issue types associated with sub-task creation
Explore the types of issues that are associated with the creation of sub-tasks in Jira. This can help you understand the different categories of problems that typically require the generation of sub-tasks.

```sql+postgres
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

```sql+sqlite
select
  id,
  name,
  description,
  avatar_id,
  subtask
from
  jira_issue_type
where
  subtask = 1;
```

### List issue types with hierarchy level 0 (Base)
Explore which issue types in a Jira project are at the base level of the hierarchy. This can be beneficial in understanding the structure of your project and identifying potential areas for reorganization.

```sql+postgres
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

```sql+sqlite
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