---
title: "Steampipe Table: jira_board - Query Jira Boards using SQL"
description: "Allows users to query Jira Boards, providing detailed insights into board configurations, types, and associated projects."
---

# Table: jira_board - Query Jira Boards using SQL

Jira Boards is a feature within Atlassian's Jira Software that allows teams to visualize their work. Boards can be customized to fit the unique workflow of any team, making it easier to manage tasks and projects. They provide a visual and interactive interface to track the progress of work.

## Table Usage Guide

The `jira_board` table provides insights into Jira Boards within Atlassian's Jira Software. As a project manager or a team lead, you can explore board-specific details through this table, including board configurations, types, and associated projects. Utilize it to uncover information about boards, such as their associated projects, the types of boards, and their configurations.

## Examples

### Basic info
Explore the types of boards in your Jira project and identify any associated filters. This can help in understanding the organization and management of tasks within your project.

```sql+postgres
select
  id,
  name,
  type,
  filter_id
from
  jira_board;
```

```sql+sqlite
select
  id,
  name,
  type,
  filter_id
from
  jira_board;
```

### List all scrum boards
Explore which project management boards are organized using the Scrum methodology. This can help you assess the prevalence and usage of this agile framework within your organization.

```sql+postgres
select
  id,
  name,
  type,
  filter_id
from
  jira_board
where
  type = 'scrum';
```

```sql+sqlite
select
  id,
  name,
  type,
  filter_id
from
  jira_board
where
  type = 'scrum';
```

### List sprints in a board
Explore the various sprints associated with a specific board to manage project timelines effectively. This can help in tracking progress and identifying any bottlenecks in the project workflow.

```sql+postgres
select
  s.board_id,
  b.name as board_name,
  b.type as board_type,
  s.id as sprint_id,
  s.name as sprint_name,
  start_date,
  end_date
from
  jira_sprint as s,
  jira_board as b
where
  s.board_id = b.id;
```

```sql+sqlite
select
  s.board_id,
  b.name as board_name,
  b.type as board_type,
  s.id as sprint_id,
  s.name as sprint_name,
  start_date,
  end_date
from
  jira_sprint as s
join
  jira_board as b
on
  s.board_id = b.id;
```