package jira

import (
	"context"
	"fmt"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableSprint(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_sprint",
		Description: "Sprint is a short period in which the development team implements and delivers a discrete and potentially shippable application increment.",
		//  TODO - Not getting board id details for get call
		// Get: &plugin.GetConfig{
		// 	KeyColumns: plugin.SingleColumn("id"),
		// 	Hydrate:    getSprint,
		// },
		List: &plugin.ListConfig{
			ParentHydrate: listBoards,
			Hydrate:       listSprints,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The ID of the sprint.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "The name of the sprint.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "board_id",
				Description: "The ID of the board the sprint belongs to.z",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("BoardId", "OriginBoardId"),
			},
			{
				Name:        "self",
				Description: "The URL of the sprint details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "Status of the sprint.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_date",
				Description: "The start timestamp of the sprint.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromCamel().NullIfZero(),
			},
			{
				Name:        "end_date",
				Description: "The projected time of completion of the sprint.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromCamel().NullIfZero(),
			},
			{
				Name:        "complete_date",
				Description: "Date the sprint was marked as complete.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromCamel().NullIfZero(),
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

func listSprints(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	board := h.Item.(jira.Board)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/agile/1.0/board/%d/sprint?startAt=%d&maxResults=%d",
			board.ID,
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
			d.StreamListItem(ctx, SprintItemInfo{int64(board.ID), sprint})
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

	if sprintId == 0 {
		return nil, nil
	}

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
	OriginBoardId int       `json:"originBoardId"`
}

type SprintItemInfo struct {
	BoardId int64
	Sprint
}
