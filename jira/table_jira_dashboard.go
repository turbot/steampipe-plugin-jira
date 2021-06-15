package jira

import (
	"context"
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableDashboard(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_dashboard",
		Description:      "Your dashboard is the main display you see when you log in to Jira.",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getDashboard,
		},
		List: &plugin.ListConfig{
			Hydrate: listDashboards,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The ID of the dashboard.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The name of the dashboard.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the dashboard details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_favourite",
				Description: "Indicates if the dashboard is selected as a favorite by the user.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "owner_account_id",
				Description: "The user info of owner of the dashboard.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Owner.AccountID"),
			},
			{
				Name:        "owner_display_name",
				Description: "The user info of owner of the dashboard.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Owner.DisplayName"),
			},
			{
				Name:        "popularity",
				Description: "The number of users who have this dashboard as a favorite.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "rank",
				Description: "The rank of this dashboard.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "view",
				Description: "The URL of the dashboard.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "edit_permissions",
				Description: "The details of any edit share permissions for the dashboard.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "share_permissions",
				Description: "The details of any view share permissions for the dashboard.",
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

func listDashboards(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/api/3/dashboard?startAt=%d&maxResults=%d",
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return nil, err
		}

		listResult := new(ListResult)
		_, err = client.Do(req, listResult)
		if err != nil {
			return nil, err
		}

		for _, dashboard := range listResult.Values {
			d.StreamListItem(ctx, dashboard)
		}

		last = listResult.StartAt + len(listResult.Values)
		if last >= listResult.Total {
			return nil, nil
		}
	}
}

//// HDRATE FUNCTIONS

func getDashboard(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	dashboardId := d.KeyColumnQuals["id"].GetStringValue()

	if dashboardId == "" {
		return nil, nil
	}

	dashboard := new(Dashboard)
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}
	apiEndpoint := fmt.Sprintf("/rest/api/3/dashboard/%s", dashboardId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	_, err = client.Do(req, dashboard)
	if err != nil && isNotFoundError(err) {
		return nil, nil
	}

	return dashboard, nil
}

type ListResult struct {
	MaxResults int         `json:"maxResults"`
	StartAt    int         `json:"startAt"`
	Total      int         `json:"total"`
	IsLast     bool        `json:"isLast"`
	Values     []Dashboard `json:"dashboards"`
}

type Dashboard struct {
	Id               string            `json:"id"`
	IsFavourite      bool              `json:"isFavourite"`
	Name             string            `json:"name"`
	Owner            jira.User         `json:"owner"`
	Popularity       int64             `json:"popularity"`
	Rank             int32             `json:"rank"`
	Self             string            `json:"self"`
	SharePermissions []SharePermission `json:"sharePermissions"`
	EditPermissions  []SharePermission `json:"editPermissions"`
	View             string            `json:"view"`
}

type SharePermission struct {
	Id   int64  `json:"id"`
	Type string `json:"type"`
}

type Owner struct {
	Id   int64  `json:"id"`
	Type string `json:"type"`
}
