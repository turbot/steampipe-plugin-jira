connection "jira" {
  plugin = "jira"

  # The baseUrl of your Jira Instance API
  # Can also be set with the JIRA_URL environment variable.
  # base_url = "https://your-domain.atlassian.net/"

  # The user name to access the jira cloud instance
  # Can also be set with the `JIRA_USER` environment variable.
  # username = "abcd@xyz.com"

  # Access Token for which to use for the API
  # Can also be set with the `JIRA_TOKEN` environment variable.
  # You should leave it empty if you are using a Personal Access Token (PAT)
  # token = "8WqcdT0rvIZpCjtDqReF48B1"

  # Personal Access Tokens are a safe alternative to using username and password for authentication.
  # This token is used in self-hosted Jira instances.
  # Can also be set with the `JIRA_PERSONAL_ACCESS_TOKEN` environment variable.
  # Personal Access Token can only be used to query jira_backlog_issue, jira_board, jira_issue and jira_sprint tables.
  # personal_access_token = "MDU0MDMx7cE25TQ3OujDfy/vkv/eeSXXoh/zXY1ex9cp"

  # Pagination size for the jira search API
  # Default is 50, max is 100.
  # page_size = 50

  # Case sensitivity
  # Default is case insensitive searches. Choose between "sensitive" and "insensitive"
  # case_sensitivity = "insensitive"

  # Issue Row Limit
  # Default is 500
  # issue_limit = 500

  # Component Row Limit
  # Default is 200
  # component_limit = 200

  # Project Row Limit
  # Default is 200
  # project_limit = 200

  # Board Row Limit
  # Default is 300
  # board_limit = 300

  # Sprint Row Limit
  # Default is 25
  # sprint_limit = 25
}
