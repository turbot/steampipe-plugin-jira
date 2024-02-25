package jira

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	jirav2 "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

type RefreshResponse struct {
	RefreshToken string `json:"refresh_token"`
}

func getStoredRefreshToken(ctx context.Context, jsonFile string) (RefreshResponse, error) {
	var refreshResponse RefreshResponse
	file, err := os.Open(jsonFile)
	if err != nil {
		plugin.Logger(ctx).Debug("Error opening ", file, " Error:", err)
		return refreshResponse, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	plugin.Logger(ctx).Debug("Loading Access token from ", jsonFile)
	err = decoder.Decode(&refreshResponse)
	if err != nil {
		plugin.Logger(ctx).Debug("Could not decode ", file, " Error:", err)
		return refreshResponse, err
	}
	plugin.Logger(ctx).Debug("Response from ", jsonFile, " ", refreshResponse)
	return refreshResponse, nil
}

func storeRefreshToken(ctx context.Context, jsonFile string, refreshResponse RefreshResponse) error {
	file, err := os.Create(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(refreshResponse)
	if err != nil {
		return err
	}
	return nil
}

func connect(ctx context.Context, d *plugin.QueryData) (*jira.Client, error) {

	// Load connection from cache, which preserves throttling protection etc
	cacheKey := "atlassian-jira"
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.(*jira.Client), nil
	}

	// Default to the env var settings
	baseUrl := os.Getenv("JIRA_URL")
	username := os.Getenv("JIRA_USER")
	token := os.Getenv("JIRA_TOKEN")
	personal_access_token := os.Getenv("JIRA_PERSONAL_ACCESS_TOKEN")
	intialRefreshToken := os.Getenv("JIRA_REFRESH_TOKEN")
	clientId := os.Getenv("JIRA_CLIENT_ID")
	clientSecret := os.Getenv("JIRA_CLIENT_SECRET")
	redirectUri := os.Getenv("OAUTH_REDIRECT_URI")
	var authMode string

	// Prefer config options given in Steampipe
	jiraConfig := GetConfig(d.Connection)

	if jiraConfig.BaseUrl != nil {
		baseUrl = *jiraConfig.BaseUrl
	}
	if jiraConfig.Username != nil {
		username = *jiraConfig.Username
	}
	if jiraConfig.Token != nil {
		token = *jiraConfig.Token
	}
	if jiraConfig.PersonalAccessToken != nil {
		personal_access_token = *jiraConfig.PersonalAccessToken
	}
	if jiraConfig.ClientId != nil {
		clientId = *jiraConfig.ClientId
	}
	if jiraConfig.ClientSecret != nil {
		clientSecret = *jiraConfig.ClientSecret
	}
	if jiraConfig.RedirectUri != nil {
		redirectUri = *jiraConfig.RedirectUri
	}
	if jiraConfig.RefreshToken != nil {
		intialRefreshToken = *jiraConfig.RefreshToken
	}

	if jiraConfig.AuthMode != nil {
		authMode = *jiraConfig.AuthMode
	} else {
		authMode = "basic"
	}

	// refresh_token = OAuth2.0(3LO)
	if baseUrl == "" {
		return nil, errors.New("'base_url' must be set in the connection configuration")
	}
	if authMode == "refresh_token" {
		if intialRefreshToken == "" {
			return nil, errors.New("refresh_token must be set in the connection configuration for OAuth2.0(3LO) flow")
		}
	} else {
		if username == "" && token != "" {
			return nil, errors.New("'token' is set but 'username' is not set in the connection configuration")
		}
		if token == "" && personal_access_token == "" && intialRefreshToken == "" {
			return nil, errors.New("'token' or 'personal_access_token' or 'refresh_token' must be set in the connection configuration")
		}
		if token != "" && personal_access_token != "" {
			return nil, errors.New("'token' and 'personal_access_token' are both set, please use only one auth method")
		}
	}

	var client *jira.Client
	var err error

	// https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/
	// curl --request POST \
	//    --url 'https://auth.atlassian.com/oauth/token' \
	//    --header 'Content-Type: application/json' \
	//    --data '{ "grant_type": "refresh_token", "client_id": "YOUR_CLIENT_ID", "client_secret": "YOUR_CLIENT_SECRET", "refresh_token": "YOUR_REFRESH_TOKEN" }'
	//
	//

	var accessToken string

	if authMode == "refresh_token" {
		refreshTokenFile := "/tmp/.jira.steampipe.7sd7sdjh324.json"
		plugin.Logger(ctx).Debug("Using Refresh token flow")
		if at, ok := d.ConnectionManager.Cache.Get("jira_access_token"); ok {
			accessToken = at.(string)
			plugin.Logger(ctx).Debug("Using cached access token")
			tokenProvider := jirav2.BearerAuthTransport{Token: accessToken}
			client, err = jira.NewClient(tokenProvider.Client(), baseUrl)
		} else {
			plugin.Logger(ctx).Debug("Access token not found in cache, fetching new access token using refresh token flow")
			var refreshToken string
			if rft, ok := d.ConnectionManager.Cache.Get("jira_refresh_token"); ok {
				plugin.Logger(ctx).Debug("Using cached refresh token", rft, ok)
				refreshToken = rft.(string)
			} else {
				plugin.Logger(ctx).Debug("Refresh token not found in cache, fetching new refresh token from store or env")
				refreshResponse, err := getStoredRefreshToken(ctx, refreshTokenFile)
				if err == nil {
					plugin.Logger(ctx).Debug("Refresh token from store")
					refreshToken = refreshResponse.RefreshToken
				} else {
					plugin.Logger(ctx).Debug("Refresh token from environment")
					refreshToken = intialRefreshToken
				}
			}
			response, err := getAccessToken(ctx, refreshToken, clientId, clientSecret, redirectUri)
			if err != nil {
				// One more try with the refresh token from the connection config
				plugin.Logger(ctx).Info("Retrying with refresh token in connection config because of ", err)
				response, err = getAccessToken(ctx, intialRefreshToken, clientId, clientSecret, redirectUri)
			}
			if err != nil {
				plugin.Logger(ctx).Error("Error getting access token: %s", err)
				return nil, fmt.Errorf("Error getting access token because of expired/invalid refresh token. : '%s'", err)
			}
			accessToken = response["access_token"].(string)
			//expiry := response["expires_in"].(int)
			refreshToken = response["refresh_token"].(string)
			d.ConnectionManager.Cache.SetWithTTL("jira_access_token", accessToken, 3000)
			d.ConnectionManager.Cache.Set("jira_access_token", refreshToken)
			plugin.Logger(ctx).Debug("Caching new access token, refresh token")
			refreshTokenResponse := RefreshResponse{RefreshToken: refreshToken}
			if err := storeRefreshToken(ctx, refreshTokenFile, refreshTokenResponse); err != nil {
				return nil, err
			}
			tokenProvider := jirav2.BearerAuthTransport{Token: accessToken}
			client, err = jira.NewClient(tokenProvider.Client(), baseUrl)
		}
	} else if personal_access_token != "" {
		// If the username is empty, let's assume the user is using a PAT
		tokenProvider := jirav2.BearerAuthTransport{Token: personal_access_token}
		client, err = jira.NewClient(tokenProvider.Client(), baseUrl)
	} else {
		plugin.Logger(ctx).Debug("Using Basic Auth flow")
		tokenProvider := jira.BasicAuthTransport{
			Username: username,
			Password: token,
		}
		client, err = jira.NewClient(tokenProvider.Client(), baseUrl)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating Jira client: %s", err.Error())
	}

	// Save to cache
	d.ConnectionManager.Cache.Set(cacheKey, client)

	// Done
	return client, nil
}

// // Constants
const (
	ColumnDescriptionTitle = "Title of the resource."
)

func getAccessToken(ctx context.Context, refreshToken string, clientId string, clientSecret string, redirectUri string) (map[string]interface{}, error) {
	// POST request to get access token and return response in JSON format
	req, err := http.NewRequest(
		"POST",
		"https://auth.atlassian.com/oauth/token",
		strings.NewReader("grant_type=refresh_token&client_id="+clientId+"&client_secret="+clientSecret+"&refresh_token="+refreshToken+"&redirect_uri="+redirectUri))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, _ := client.Do(req)
	if resp.StatusCode != 200 {
		plugin.Logger(ctx).Error("Error: ", resp.Status)
		return nil, fmt.Errorf("Error: %s", resp.Status)
	}
	response := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

//// TRANSFORM FUNCTION

// convertJiraTime:: converts jira.Time to time.Time
func convertJiraTime(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	if v, ok := d.Value.(jira.Time); ok {
		return time.Time(v), nil
	} else if v, ok := d.Value.(*jira.Time); ok {
		return time.Time(*v), nil
	}
	return nil, nil
}

// convertJiraDate:: converts jira.Date to time.Time
func convertJiraDate(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	return time.Time(d.Value.(jira.Date)), nil
}

func buildJQLQueryFromQuals(ctx context.Context, equalQuals plugin.KeyColumnQualMap, tableColumns []*plugin.Column) string {
	filters := []string{}
	plugin.Logger(ctx).Debug("jira_issue::buildJQLQueryFromQuals", equalQuals)
	for _, filterQualItem := range tableColumns {
		filterQual := equalQuals[filterQualItem.Name]

		if filterQual == nil {
			continue
		}

		// Check only if filter qual map matches with optional column name
		if filterQual.Name == filterQualItem.Name {
			if filterQual.Quals == nil {
				continue
			}

			for _, qual := range filterQual.Quals {
				if qual.Value != nil {
					value := qual.Value
					//plugin.Logger(ctx).Debug("jira_issue::buildJQLQueryFromQuals", value, filterQualItem.Type, qual.Operator)

					switch filterQualItem.Type {
					case proto.ColumnType_STRING:
						switch qual.Operator {
						case "=":
							filters = append(filters, fmt.Sprintf("\"%s\" = \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
						case "<>":
							filters = append(filters, fmt.Sprintf("\"%s\" != \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
							//case "~~":
							//	filters = append(filters, fmt.Sprintf("%s = \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetStringValue()))
						}
					case proto.ColumnType_TIMESTAMP:
						switch qual.Operator {
						case "=", ">=", ">", "<=", "<":
							filters = append(filters, fmt.Sprintf("\"%s\" %s \"%s\"", getIssueJQLKey(filterQualItem.Name), qual.Operator, value.GetTimestampValue().AsTime().Format("2006-01-02 15:04")))
						case "<>":
							filters = append(filters, fmt.Sprintf("\"%s\" != \"%s\"", getIssueJQLKey(filterQualItem.Name), value.GetTimestampValue().AsTime().Format("2006-01-02 15:04")))
						}

					}
				}
			}

		}
	}

	if len(filters) > 0 {
		return strings.Join(filters, " AND ")
	}

	return ""
}

func getRequiredCustomField() map[string]map[string]interface{} {
	customFieldMap := map[string]map[string]interface{}{
		"Eng Target Version/s": {
			"key":        "customfield_13193",
			"name":       "Eng Target Version/s",
			"searchable": true,
			"type":       "array",
		},
		"Release Commit": {
			"key":        "customfield_13139",
			"name":       "Release Commit",
			"searchable": true,
			"type":       "option",
		},
		"V-team/P-team": {
			"key":        "customfield_13323",
			"name":       "V-team/P-team",
			"searchable": true,
			"type":       "option-with-child",
		},
		"Found-In Version": {
			"key":        "customfield_13149",
			"name":       "Found-In Version",
			"searchable": true,
			"type":       "option-with-child",
		},
		"sprint": {
			"key":        "customfield_10007",
			"name":       "Sprint",
			"searchable": true,
			"type":       "array",
		},
		"epic": {
			"key":        "customfield_10300",
			"name":       "Epic Link",
			"searchable": true,
			"type":       "any",
		},
	}
	return customFieldMap
}

func getIssueJQLKey(columnName string) string {
	customFieldMap := getRequiredCustomField()
	jqlFieldMap := map[string]string{
		"resolution_date":        "resolutionDate",
		"status_category":        "statusCategory",
		"parent_key":             "parent",
		"parent_status":          "parentStatus",
		"parent_status_category": "parentStatusCategory",
		"parent_issue_type":      "parentIssueType",
	}
	// if the column name is in the map, return the value else return the column name
	if field, ok := jqlFieldMap[columnName]; ok {
		return field
	} else if customField, ok := customFieldMap[columnName]; ok {
		return customField["name"].(string)
	} else {
		return strings.ToLower(strings.Split(columnName, "_")[0])
	}
}

func getPageSize(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	pageSize := 50
	if jiraConfig.PageSize != nil {
		pageSize = *jiraConfig.PageSize
	}

	if pageSize < 1 || pageSize > 100 {
		return -1, errors.New("'page_size' must be set to 1 to 100 in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	return pageSize, nil
}

func getCaseSensitivity(_ context.Context, d *plugin.QueryData) (string, error) {
	jiraConfig := GetConfig(d.Connection)

	caseSensitivity := "insensitive"
	if jiraConfig.CaseSensitivity != nil {
		caseSensitivity = *jiraConfig.CaseSensitivity
	}

	if caseSensitivity != "sensitive" && caseSensitivity != "insensitive" {
		return "", errors.New("'case_sensitive' must be set to 'insensitive' or 'sensitive' in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	return caseSensitivity, nil
}

func getIssueLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	issueLimit := 500
	if jiraConfig.IssueLimit != nil {
		issueLimit = *jiraConfig.IssueLimit
	}

	if issueLimit < 1 {
		return -1, errors.New("'issue_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return issueLimit, nil
}

func getComponentLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	componentLimit := 200
	if jiraConfig.ComponentLimit != nil {
		componentLimit = *jiraConfig.ComponentLimit
	}

	if componentLimit < 1 {
		return -1, errors.New("'component_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return componentLimit, nil
}

func getProjectLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	projectLimit := 200
	if jiraConfig.ProjectLimit != nil {
		projectLimit = *jiraConfig.ProjectLimit
	}

	if projectLimit < 1 {
		return -1, errors.New("'project_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return projectLimit, nil
}

func getBoardLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	boardLimit := 300
	if jiraConfig.BoardLimit != nil {
		boardLimit = *jiraConfig.BoardLimit
	}

	if boardLimit < 1 {
		return -1, errors.New("'board_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return boardLimit, nil
}

func getSprintLimit(_ context.Context, d *plugin.QueryData) (int, error) {
	jiraConfig := GetConfig(d.Connection)

	sprintLimit := 25
	if jiraConfig.SprintLimit != nil {
		sprintLimit = *jiraConfig.SprintLimit
	}

	if sprintLimit < 1 {
		return -1, errors.New("'sprint_limit' must be greater than 0. Edit your connection configuration file and then restart Steampipe")
	}

	return sprintLimit, nil
}

func getRowLimitError(_ context.Context, d *plugin.QueryData) (bool, error) {
	jiraConfig := GetConfig(d.Connection)

	rowLimitError := true
	if jiraConfig.RowLimitError != nil {
		rowLimitError = *jiraConfig.RowLimitError
	}

	return rowLimitError, nil
}
