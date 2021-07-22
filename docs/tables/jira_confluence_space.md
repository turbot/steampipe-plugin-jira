# Table: jira_confluence_space

**Confluence Spaces** are collections of related pages that you and other people in your team or organization work on together. Most organizations use a mix of team spaces, software project spaces, documentation spaces, and knowledge base spaces.

## Examples

### Basic info

```sql
select
  id,
  key,
  name,
  status,
  type
from
  jira_confluence_space;
```

### List all archived spaces

```sql
select
  id,
  key,
  name,
  status,
  type
from
  jira_confluence_space
where
  status = 'archived';
```

### List spaces without any description

```sql
select
  id,
  key,
  name,
  description,
  status,
  type
from
  jira_confluence_space
where
  description = '';
```

### List spaces belongs to global type

```sql
select
  id,
  key,
  name,
  status,
  type
from
  jira_confluence_space
where
  type = 'global';
```

### List spaces belongs to knowledge-bases category

```sql
select
  id,
  key,
  name,
  status,
  type,
  category
from
  jira_confluence_space
where
  category @> '"knowledge-bases"';
```
