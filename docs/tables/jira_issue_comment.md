# Table: jira_issue_comment

Comments in Jira are an integral part of issue management, ensuring transparent and effective communication among team members and stakeholders, which is essential for successful project completion and issue resolution.

## Examples

### Basic info

```sql
select
  id,
  self,
  issue_id,
  author
from
  jira_issue_comment;
```

### List comments that are hidden in Service Desk

```sql
select
  id,
  self,
  issue_id,
  body,
  created,
  jsd_public
from
  jira_issue_comment
where
  jsd_public;
```

### List comments posted in the last 5 days for a particular issues

```sql
select
  id,
  issue_id,
  body,
  created
from
  jira_issue_comment
where
  created >= now() - interval '5' day
  and issue_id = '10021';
```

### List comments that updated in lsdt 2 hours

```sql
select
  id,
  issue_id,
  body,
  created,
  updated
from
  jira_issue_comment
where
  updated >= now() - interval '2' hour;
```

### Get author information of comments

```sql
select
  id,
  issue_id,
  author ->> 'accountId' as author_account_id,
  author ->> 'accountType' as author_account_type,
  author ->> 'displayName' as author_name,
  author ->> 'emailAddress' as author_email_address,
  author ->> 'timeZone' as author_time_zone
from
  jira_issue_comment;
```
