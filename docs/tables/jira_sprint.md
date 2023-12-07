---
title: "Steampipe Table: jira_sprint - Query Jira Sprints using SQL"
description: "Allows users to query Jira Sprints, providing insights into the progress, status, and details of each sprint within a Jira project."
---

# Table: jira_sprint - Query Jira Sprints using SQL

Jira Sprints are a key component of the agile project management offered by Jira. They represent a set timeframe during which specific work has to be completed and made ready for review. Sprints are used in the Scrum framework and they help teams to better manage and organize their work, by breaking down larger projects into manageable, time-boxed periods.

## Table Usage Guide

The `jira_sprint` table provides insights into the sprints within a Jira project. As a project manager or a team lead, you can explore details about each sprint, including its progress, status, and associated issues. Use this table to track the progress of ongoing sprints, plan for future sprints, and review the performance of past sprints.

## Examples

### Basic info
Explore which projects are currently active, along with their respective timelines. This can assist in understanding the progress and schedules of different projects.

```sql+postgres
select
  id,
  name,
  board_id,
  state,
  start_date,
  end_date,
  complete_date
from
  jira_sprint;
```

```sql+sqlite
select
  id,
  name,
  board_id,
  state,
  start_date,
  end_date,
  complete_date
from
  jira_sprint;
```

### List sprints due in the next week
Explore which sprints are due in the coming week. This can help in planning and prioritizing tasks accordingly.

```sql+postgres
select
  id,
  name,
  board_id,
  state,
  start_date,
  end_date
from
  jira_sprint
where
  end_date > current_date
  and end_date <= (current_date + interval '7' day);
```

```sql+sqlite
select
  id,
  name,
  board_id,
  state,
  start_date,
  end_date
from
  jira_sprint
where
  end_date > date('now')
  and end_date <= date('now', '+7 day');
```

### List active sprints
Explore which sprints are currently active in your Jira project. This query helps in gaining insights into ongoing tasks and managing project timelines effectively.

```sql+postgres
select
  id,
  name,
  board_id,
  start_date,
  end_date
from
  jira_sprint
where
  state = 'active';
```

```sql+sqlite
select
  id,
  name,
  board_id,
  start_date,
  end_date
from
  jira_sprint
where
  state = 'active';
```

### List issues in a sprints
Discover the segments that require attention by identifying active tasks within a particular project phase, helping to allocate resources more effectively and manage project timelines.

```sql+postgres
select
  id,
  key,
  summary,
  project_key,
  status,
  assignee_display_name,
  assignee_account_id,
  duedate
from
  jira_issue
where
  sprint_ids @> '2';
```

```sql+sqlite
Error: SQLite does not support array contains operator '@>'.
```

### Count of issues by sprint
This query is useful for understanding the distribution of tasks across different sprints in a project. It provides an overview of workload allocation, helping to identify if certain sprints are overloaded with issues.
```sql+postgres
select 
  b.name as board_name,
  s.name as sprint_name,
  count(sprint_id) as num_issues
from 
  jira_board as b
  join jira_sprint as s on s.board_id = b.id 
  left join jira_issue as i on true 
  left join jsonb_array_elements(i.sprint_ids) as sprint_id on sprint_id ::bigint = s.id  
group by
  board_name,
  sprint_name
order by
  board_name,
  sprint_name;
```

```sql+sqlite
select 
  b.name as board_name,
  s.name as sprint_name,
  count(sprint_id) as num_issues
from 
  jira_board as b
  join jira_sprint as s on s.board_id = b.id 
  left join jira_issue as i 
  left join json_each(i.sprint_ids) as sprint_id on sprint_id.value = s.id  
group by
  board_name,
  sprint_name
order by
  board_name,
  sprint_name;
```