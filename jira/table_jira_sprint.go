package jira

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableSprint(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_sprint",
		Description:      "Jira Sprint",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getSprint,
		},
		List: &plugin.ListConfig{
			KeyColumns: plugin.SingleColumn("board_id"),
			Hydrate:    listSprints,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "A friendly name that identifies the board.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "board_id",
				Description: "The unique identifier of board.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "self",
				Description: "The unique identifier of board.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The unique identifier of board.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_date",
				Description: "The unique identifier of board.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromCamel().NullIfZero(),
			},
			{
				Name:        "end_date",
				Description: "The unique identifier of board.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromCamel().NullIfZero(),
			},
			{
				Name:        "complete_date",
				Description: "The unique identifier of board.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromCamel().NullIfZero(),
			},
			{
				Name:      "origin_board_id",
				Type:      proto.ColumnType_INT,
				Transform: transform.FromCamel().NullIfZero(),
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

func listSprints(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	boardId := d.KeyColumnQuals["board_id"].GetInt64Value()

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/agile/1.0/board/%d/sprint?startAt=%d&maxResults=%d",
			boardId,
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return nil, err
		}

		listResult := new(ListSprintResult)
		_, err = client.Do(req, listResult)
		if err != nil && isNotFoundError(err) {
			if isNotFoundError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, sprint := range listResult.Values {
			d.StreamListItem(ctx, SprintItemInfo{boardId, sprint})
		}

		last = listResult.StartAt + len(listResult.Values)
		if listResult.IsLast {
			return nil, nil
		}
	}
}

//// HDRATE FUNCTIONS

func getSprint(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	sprintId := d.KeyColumnQuals["id"].GetInt64Value()

	sprint := new(Sprint)
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	_, err = client.Do(req, sprint)
	if err != nil && isNotFoundError(err) {
		return nil, nil
	}

	return sprint, err
}

type ListSprintResult struct {
	MaxResults int      `json:"maxResults"`
	StartAt    int      `json:"startAt"`
	Total      int      `json:"total"`
	IsLast     bool     `json:"isLast"`
	Values     []Sprint `json:"values"`
}

type Sprint struct {
	Id            int64     `json:"id"`
	Self          string    `json:"self"`
	Name          string    `json:"name"`
	State         string    `json:"state"`
	EndDate       time.Time `json:"endDate"`
	StartDate     time.Time `json:"startDate"`
	CompleteDate  time.Time `json:"completeDate"`
	OriginBoardId int64     `json:"originBoardId"`
}

type SprintItemInfo struct {
	BoardId int64
	Sprint
}
