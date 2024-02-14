---
title: "Steampipe Table: jira_workflow - Query Jira Workflows using SQL"
description: "Allows users to query Jira Workflows, providing insights into the steps, transitions, and status categories of each workflow."
---

# Table: jira_workflow - Query Jira Workflows using SQL

Jira Workflows is a feature within Atlassian's Jira software that enables teams to manage and track the lifecycle of tasks and issues. It provides a visual representation of the process an issue goes through from creation to completion, allowing teams to customize and control how their work flows. Jira Workflows helps in determining the steps an issue needs to go through to reach resolution, setting permissions for who can move issues between steps, and automating these transitions.

## Table Usage Guide

The `jira_workflow` table provides a detailed view of Jira Workflows within a Jira software instance. As a project manager or a team lead, leverage this table to gain insights into the steps, transitions, and status categories of each workflow. Utilize it to manage and optimize your team's work process, understand the lifecycle of tasks, and identify bottlenecks in your project's workflow.

## Examples

### Basic info
Analyze the settings to understand the default workflows in your Jira system, enabling you to better manage your project processes and prioritize tasks. This is particularly useful for project managers who need to assess the elements within their workflows and identify areas for improvement or customization.

```sql+postgres
select
  name,
  entity_id,
  description,
  is_default
from
  jira_workflow;
```

```sql+sqlite
select
  name,
  entity_id,
  description,
  is_default
from
  jira_workflow;
```

### List workflows that are not default
Uncover the details of workflows in Jira that have been customized and are not set as default. This can be beneficial for administrators to understand the unique workflows in their system and make necessary adjustments or improvements.

```sql+postgres
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

```sql+sqlite
select
  name,
  entity_id,
  description,
  is_default
from
  jira_workflow
where
  is_default = 0;
```

### List workflows that are not associated with entity id
Discover workflows that lack an associated entity ID, which could indicate incomplete or misconfigured processes within your Jira workflow system. This can be useful to identify and rectify potential issues, ensuring smoother operations.

```sql+postgres
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

```sql+sqlite
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