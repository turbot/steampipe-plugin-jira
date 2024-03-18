package jira

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableFilter(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_filter",
		Description: "List all filters. ",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getFilter,
		},
		List: &plugin.ListConfig{
			Hydrate: listFilters,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the filter.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The ID of the filter.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "description",
				Description: "Description of the filter.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "jql",
				Description: "JQL used for this filter.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "owner_account_id",
				Description: "Account Id of the owner of filter.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Owner.AccountID"),
			},
			{
				Name:        "owner_display_name",
				Description: "Display name of the owner of filter.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Owner.DisplayName"),
			},
			// Steampipe standard columns
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

func listFilters(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	maxResults := 100
	startAt := 0
	for {
		filters, err := searchForFilters(ctx, d, startAt, maxResults)
		if err != nil {
			return nil, err
		}
		for _, filter := range (*filters).Filters {
			d.StreamListItem(ctx, filter)
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if err != nil {
			return nil, err
		}
		startAt += maxResults
		if startAt >= filters.Total {
			break
		}
	}

	return nil, nil
}

func searchForFilters(ctx context.Context, d *plugin.QueryData, startAt int, maxResults int) (*filterSearchResult, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_filter.searchForFilters", "connection_error", err)
		return nil, err
	}

	u := url.URL{
		Path: "rest/api/3/filter/search",
	}
	uv := url.Values{}

	// Append the values of options to the path parameters

	uv.Add("startAt", strconv.Itoa(startAt))
	uv.Add("maxResults", strconv.Itoa(maxResults))
	uv.Add("expand", "jql,description,owner")
	u.RawQuery = uv.Encode()

	req, err := client.NewRequest("GET", u.String(), nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_filter.searchForFilters", "get_request_error", err)
		return nil, err
	}
	filters := new(filterSearchResult)

	res, err := client.Do(req, filters)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_filter.searchForFilters", "res_body", string(body))

	if err != nil {
		plugin.Logger(ctx).Error("jira_filter.searchForFilters", "api_error", err)
		return nil, err
	}
	return filters, nil
}

//// HYDRATE FUNCTIONS

func getFilter(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get this entire table from cache
	// and filter this by key

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_filter.getFilter", "connection_error", err)
		return nil, err
	}
	filterId := d.EqualsQuals["id"].GetStringValue()

	// Return nil, if no input provided
	if filterId == "" {
		return nil, nil
	}

	apiEndpoint := fmt.Sprintf("rest/api/3/filter/%s", filterId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_filter.getFilter", "get_request_error", err)
		return nil, err
	}
	result := new(filterSearchResult)

	res, err := client.Do(req, result)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_filter.getFilter", "res_body", string(body))

	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_filter.getFilter", "api_error", err)
		return nil, err
	}

	return result, nil
}

type filterSearchResult struct {
	Filters    []jira.Filter `json:"values" structs:"values"`
	StartAt    int           `json:"startAt" structs:"startAt"`
	MaxResults int           `json:"maxResults" structs:"maxResults"`
	Total      int           `json:"total" structs:"total"`
}
