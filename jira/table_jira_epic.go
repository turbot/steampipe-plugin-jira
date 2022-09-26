package jira

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v3/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v3/plugin"
)

//// TABLE DEFINITION

func tableEpic(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_epic",
		Description: "An epic is essentially a large user story that can be broken down into a number of smaller stories. An epic can span more than one project.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AnyColumn([]string{"id", "key"}),
			Hydrate:    getEpic,
		},
		List: &plugin.ListConfig{
			Hydrate: listEpics,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The id of the epic.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "The name of the epic.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "key",
				Description: "The key of the epic.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "done",
				Description: "Indicates the status of the epic.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "self",
				Description: "The URL of the epic details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "summary",
				Description: "Description of the epic.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "color",
				Description: "Label colour details for the epic.",
				Type:        proto.ColumnType_JSON,
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		},
	}
}

//// LIST FUNCTION

func listEpics(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_epic.listEpics", "connection_error", err)
		return nil, err
	}

	last := 0
	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var maxResults int = 100
	if d.QueryContext.Limit != nil {
		if *queryLimit < 100 {
			maxResults = int(*queryLimit)
		}
	}
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/agile/1.0/epic/search?startAt=%d&maxResults=%d",
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_epic.listEpics", "get_request_error", err)
			return nil, err
		}

		listResult := new(ListEpicResult)
		res, err := client.Do(req, listResult)
		body, _ := ioutil.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_epic.listEpics", "res_body", string(body))
		if err != nil {
			plugin.Logger(ctx).Error("jira_epic.listEpics", "api_error", err)
			return nil, err
		}

		for _, epic := range listResult.Values {
			d.StreamListItem(ctx, epic)
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.QueryStatus.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		last = listResult.StartAt + len(listResult.Values)
		if listResult.IsLast {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getEpic(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getEpic")

	epicId := d.KeyColumnQuals["id"].GetInt64Value()
	epicKey := d.KeyColumnQuals["key"].GetStringValue()
	var apiEndpoint string

	if epicKey != "" {
		apiEndpoint = fmt.Sprintf("/rest/agile/1.0/epic/%s", epicKey)
	} else {
		apiEndpoint = fmt.Sprintf("/rest/agile/1.0/epic/%d", epicId)
	}

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_epic.getEpic", "connection_error", err)
		return nil, err
	}

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_epic.getEpic", "get_request_error", err)
		return nil, err
	}

	epic := new(Epic)
	res, err := client.Do(req, epic)
	body, _ := ioutil.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_epic.getEpic", "res_body", string(body))
	if err != nil {
		if isNotFoundError(err) || strings.Contains(err.Error(), "400") {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_epic.getEpic", "api_error", err)
		return nil, err
	}

	return epic, err
}

//// Custom Structs

type ListEpicResult struct {
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"`
	Total      int    `json:"total"`
	IsLast     bool   `json:"isLast"`
	Values     []Epic `json:"values"`
}

type Epic struct {
	Color   Color  `json:"color"`
	Done    bool   `json:"done"`
	Id      int64  `json:"id"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Self    string `json:"self"`
	Summary string `json:"summary"`
}

type Color struct {
	Key string `json:"key"`
}
