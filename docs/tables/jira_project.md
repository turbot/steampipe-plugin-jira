---
title: "Steampipe Table: jira_project - Query Jira Projects using SQL"
description: "Allows users to query Jira Projects, offering detailed insights into project-specific information such as project key, project type, project lead, and project category."
---

# Table: jira_project - Query Jira Projects using SQL

Jira is a popular project management tool used by software development teams to plan, track, and release software. It offers a range of features including bug tracking, issue tracking, and project management functionalities. Jira Projects are the primary containers in Jira where issues are created and tracked.

## Table Usage Guide

The `jira_project` table provides insights into Projects within Jira. As a Project Manager or a Scrum Master, you can explore project-specific details through this table, including project key, project type, project lead, and project category. Use it to uncover information about projects, such as the project's status, the lead responsible for the project, and the category the project belongs to.

## Examples

### Basic info
Explore the different projects within your Jira environment, gaining insights into key aspects like the project's name, ID, key, lead display name, category, and description. This can be useful for understanding the scope and management of your projects.

```sql+postgres
select
  name,
  id,
  key,
  lead_display_name,
  project_category,
  description
from
  jira_project;
```

```sql+sqlite
select
  name,
  id,
  key,
  lead_display_name,
  project_category,
  description
from
  jira_project;
```

### List all issues in a project
Explore the status and details of all issues within a specific project. This can be useful for project management, allowing you to assess the workload and track the progress of tasks.

```sql+postgres
select
  id,
  key,
  project_id,
  project_key,
  created,
  creator_display_name,
  assignee_display_name,
  status,
  summary
from
  jira_issue
where
  project_key = 'TEST';
```

```sql+sqlite
select
  id,
  key,
  project_id,
  project_key,
  created,
  creator_display_name,
  assignee_display_name,
  status,
  summary
from
  jira_issue
where
  project_key = 'TEST';
```