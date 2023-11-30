---
title: "Steampipe Table: jira_group - Query Jira Groups using SQL"
description: "Allows users to query Jira Groups, providing insights into the various user groups within a Jira instance."
---

# Table: jira_group - Query Jira Groups using SQL

Jira Groups are a collection of users within an instance of Jira, a popular project management tool. Groups offer an efficient way to manage a collection of users. They can be used to set permissions, restrict access, and manage notifications across a Jira instance.

## Table Usage Guide

The `jira_group` table provides insights into the user groups within a Jira instance. As a project manager or system administrator, you can explore group-specific details through this table, including group names, users within each group, and associated permissions. Utilize it to manage access control, set permissions, and streamline user management within your Jira instance.

## Examples

### Basic info
Explore the different groups present in your Jira setup, allowing you to manage and organize users more effectively. This can be particularly useful when you need to assign tasks to specific groups or manage permissions.

```sql
select
  name,
  id
from
  jira_group;
```

### Get associated users
Discover the segments that are associated with specific users in a group. This can help in managing user access and permissions effectively.

```sql
select
  name as group_name,
  u.display_name as user_name,
  user_id,
  u.email_address as user_email
from
  jira_group as g,
  jsonb_array_elements_text(member_ids) as user_id,
  jira_user as u
where
  user_id = u.account_id
order by
  group_name,
  user_name;
```