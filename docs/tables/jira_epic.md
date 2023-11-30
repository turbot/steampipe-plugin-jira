---
title: "Steampipe Table: jira_epic - Query Jira Epics using SQL"
description: "Allows users to query Jira Epics, providing insights into the progress, status, and details of each epic within a Jira project."
---

# Table: jira_epic - Query Jira Epics using SQL

Jira is a project management tool developed by Atlassian. It provides a platform for planning, tracking, and managing agile software development projects. An Epic in Jira is a large user story that can be broken down into several smaller stories.

## Table Usage Guide

The `jira_epic` table provides insights into Epics within Jira. As a project manager or a team lead, explore Epic-specific details through this table, including progress, status, and associated tasks. Utilize it to uncover information about Epics, such as those that are overdue, the relationship between tasks and Epics, and the overall progress of a project.

## Examples

### Basic info
Explore the status and summary of various tasks within a project management tool to understand their progression and key details. This can help in monitoring progress and identifying any bottlenecks or issues that need to be addressed.

```sql
select
  id,
  name,
  key,
  done as status,
  summary
from
  jira_epic;
```

### List issues in epic
Explore which tasks are associated with which project milestones by identifying instances where issue status, creation date, and assignee are linked with a specific project epic. This can help in assessing the distribution and management of tasks across different project stages.

```sql
select
  i.id as issue_id,
  i.status as issue_status,
  i.created as issue_created,
  i.creator_display_name,
  i.assignee_display_name,
  e.id as epic_id,
  e.name as epic_name,
  i.summary as issue_summary
from
  jira_epic as e,
  jira_issue as i
where
  i.epic_key = e.key;
```