---
title: "Steampipe Table: jira_component - Query Jira Components using SQL"
description: "Allows users to query Jira Components, specifically to retrieve details about individual components within a Jira project, providing insights into component name, description, lead details, and project keys."
---

# Table: jira_component - Query Jira Components using SQL

A Jira Component is a subsection of a project. They are used to group issues within a project into smaller parts. You can set a default assignee for a component, which will override the project's default assignee.

## Table Usage Guide

The `jira_component` table provides insights into the components within a Jira project. As a Project Manager or Developer, explore component-specific details through this table, including component name, description, lead details, and project keys. Utilize it to manage and organize issues within a project, making project management more efficient and streamlined.

## Examples

### Basic info
Explore which components of a project have the most issues, helping to identify areas that may need additional resources or attention.

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
Determine the areas in which components are experiencing issues, allowing you to assess and address problem areas within your projects effectively.

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
Determine the areas in your project where components lack assigned leads. This can help in identifying potential bottlenecks and ensuring responsibilities are properly delegated.

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