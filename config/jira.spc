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

  # Personal Access Token for which to use for the API.
  # This one isused in self-hosted Jira instances.
  # Can also be set with the `JIRA_PERSONAL_ACCESS_TOKEN` environment variable.
  # personal_access_token = "MDU0MDMx7cE25TQ3OujDfy/vkv/eeSXXoh/zXY1ex9cp"
}
