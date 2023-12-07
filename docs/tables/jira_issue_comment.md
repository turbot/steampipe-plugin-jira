---
title: "Steampipe Table: jira_issue_comment - Query Jira Issue Comments using SQL"
description: "Allows users to query Jira Issue Comments, providing insights into the details of comments made on issues within the Jira platform."
---

# Table: jira_issue_comment - Query Jira Issue Comments using SQL

Jira is a project management tool that aids in issue tracking and agile project management. It is widely used by software development teams for planning, tracking progress, and releasing software. Issue comments in Jira serve as a communication medium on the platform, allowing users to discuss, provide updates, and track changes regarding specific issues.

## Table Usage Guide

The `jira_issue_comment` table provides insights into the comments made on issues within the Jira platform. As a project manager or a team member, you can explore comment-specific details through this table, including the author of the comment, creation date, and the issue the comment is associated with. Utilize it to track communication, understand the context of discussions, and monitor the progress of issues.

## Examples

### Basic info
Explore the comments on different issues in Jira by identifying their unique identifiers and authors. This can be useful in understanding who is actively participating in discussions and contributing to issue resolutions.

```sql+postgres
select
  id,
  self,
  issue_id,
  author
from
  jira_issue_comment;
```

```sql+sqlite
select
  id,
  self,
  issue_id,
  author
from
  jira_issue_comment;
```

### List comments that are hidden in Service Desk
Discover the segments that contain hidden comments in the Service Desk to gain insights into user feedback or issues that may not be publicly visible. This can be useful in understanding customer concerns and improving service quality.

```sql+postgres
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

```sql+sqlite
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
Explore recent feedback or updates on a specific issue by identifying comments posted in the last five days. This can help in tracking the progress of issue resolution and understanding the current discussion around it.

```sql+postgres
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

```sql+sqlite
select
  id,
  issue_id,
  body,
  created
from
  jira_issue_comment
where
  created >= datetime('now', '-5 days')
  and issue_id = '10021';
```

### List comments that were updated in last 2 hours
Explore recent activity by identifying comments that have been updated in the past two hours. This is useful for staying informed about ongoing discussions or changes in your Jira issues.

```sql+postgres
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

```sql+sqlite
select
  id,
  issue_id,
  body,
  created,
  updated
from
  jira_issue_comment
where
  updated >= datetime('now', '-2 hours');
```

### Get author information of a particular issue comment
Explore the identity of the individual who commented on a specific issue. This can be beneficial for understanding who is contributing to the discussion or if any particular individual's input requires further attention.

```sql+postgres
select
  id,
  issue_id,
  author ->> 'accountId' as author_account_id,
  author ->> 'accountType' as author_account_type,
  author ->> 'displayName' as author_name,
  author ->> 'emailAddress' as author_email_address,
  author ->> 'timeZone' as author_time_zone
from
  jira_issue_comment
where
  id = '10015';
```

```sql+sqlite
select
  id,
  issue_id,
  json_extract(author, '$.accountId') as author_account_id,
  json_extract(author, '$.accountType') as author_account_type,
  json_extract(author, '$.displayName') as author_name,
  json_extract(author, '$.emailAddress') as author_email_address,
  json_extract(author, '$.timeZone') as author_time_zone
from
  jira_issue_comment
where
  id = '10015';
```