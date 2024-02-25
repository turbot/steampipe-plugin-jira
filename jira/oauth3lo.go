package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type OAuth3LOConfig struct {
	ClientId     string
	ClientSecret string
	RedirectUri  string
	RefreshToken string
	RefreshUri   string
	TokenFile    string
}
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

func oauthRequest(ctx context.Context, refreshToken string, cfg OAuth3LOConfig) (map[string]interface{}, error) {
	// POST request to get access token and return response in JSON format
	req, err := http.NewRequest(
		"POST",
		"https://auth.atlassian.com/oauth/token",
		strings.NewReader("grant_type=refresh_token&client_id="+cfg.ClientId+"&client_secret="+cfg.ClientSecret+"&refresh_token="+refreshToken+"&redirect_uri="+cfg.RedirectUri))
	if err != nil {
		return nil, err
	}
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

func getRefreshToken(ctx context.Context, d *plugin.QueryData, cfg OAuth3LOConfig) (string, error) {
	// POST request to get access token and return response in JSON format
	var refreshToken string

	if rft, ok := d.ConnectionManager.Cache.Get("jira_refresh_token"); ok {
		plugin.Logger(ctx).Debug("Using cached Refresh token", rft, ok)
		refreshToken = rft.(string)
	} else {
		plugin.Logger(ctx).Debug("Refresh token not found in cache, fetching new refresh token from store or env")
		refreshResponse, err := getStoredRefreshToken(ctx, cfg.TokenFile)
		if err == nil {
			plugin.Logger(ctx).Debug("Refresh token from store")
			refreshToken = refreshResponse.RefreshToken
		} else {
			plugin.Logger(ctx).Debug("Refresh token from environment")
			refreshToken = cfg.RefreshToken
		}
	}
	return refreshToken, nil
}

func getAccessToken(ctx context.Context, d *plugin.QueryData, cfg OAuth3LOConfig) (string, *time.Duration, error) {

	plugin.Logger(ctx).Debug("Using Refresh token flow")
	var ttl *time.Duration = nil
	var accessToken string
	if at, ok := d.ConnectionManager.Cache.Get("jira_access_token"); ok {
		accessToken = at.(string)
		plugin.Logger(ctx).Debug("Using cached access token")
	} else {
		plugin.Logger(ctx).Debug("Access token not found in cache, fetching new access token using refresh token flow")
		refreshToken, err := getRefreshToken(ctx, d, cfg)
		if err != nil {
			return "", nil, fmt.Errorf("No Refresh Token found. : '%s'", err)
		}

		response, e := oauthRequest(ctx, refreshToken, cfg)
		if e != nil {
			// One more try with the refresh token from the connection config
			plugin.Logger(ctx).Info("Retrying with refresh token in connection config because of ", e)
			response, e = oauthRequest(ctx, cfg.RefreshToken, cfg)
		}
		if e != nil {
			plugin.Logger(ctx).Error("Error getting access token: %s", e)
			return "", nil, fmt.Errorf("Error getting access token because of expired/invalid refresh token. : '%s'", e)
		}
		accessToken = response["access_token"].(string)
		expiry := 3000 * time.Second
		if response["expires_in"] != nil {
			expiry = time.Duration(response["expires_in"].(float64)) * time.Second
			if expiry > 60*time.Second {
				expiry = expiry - 60*time.Second
			}
		}
		ttl = &expiry
		refreshToken = response["refresh_token"].(string)
		plugin.Logger(ctx).Error("Setting Token Expiry time after ", expiry)
		d.ConnectionManager.Cache.SetWithTTL("jira_access_token", accessToken, expiry)
		d.ConnectionManager.Cache.Set("jira_access_token", refreshToken)
		plugin.Logger(ctx).Debug("Caching new access token, refresh token")
		refreshTokenResponse := RefreshResponse{RefreshToken: refreshToken}
		if storeError := storeRefreshToken(ctx, cfg.TokenFile, refreshTokenResponse); e != nil {
			return "", nil, storeError
		}
	}
	return accessToken, ttl, nil
}
