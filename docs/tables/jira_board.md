# Table: jira_board

A board displays issues from one or more projects, giving you a flexible way of viewing, managing, and reporting on work in progress.
There are three types of boards in Jira Software:

- Team-managed board: For teams who are new to agile. Get your team up-and-running with this simplified board. The set-up is straight-forward and streamlined, delivering more power progressively as you need it.

- Scrum board: For teams that plan their work in sprints. Includes a backlog.

- Kanban board: For teams that focus on managing and constraining their work-in-progress. Includes the option of a Kanban backlog.

## Examples

### Basic info

```sql
select
  id,
  name,
  type,
  filter_id
from
  jira_board;
```

### List all scrum boards

```sql
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

```sql
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
