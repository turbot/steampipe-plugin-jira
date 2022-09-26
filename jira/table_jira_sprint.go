package jira

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v3/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v3/plugin"
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
		plugin.Logger(ctx).Error("jira_sprint.listSprints", "connection_error", err)
		return nil, err
	}

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var maxResults int = 1000
	if d.QueryContext.Limit != nil {
		if *queryLimit < 1000 {
			maxResults = int(*queryLimit)
		}
	}

	last := 0
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/agile/1.0/board/%d/sprint?startAt=%d&maxResults=%d",
			board.ID,
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_sprint.listSprints", "get_request_error", err)
			return nil, err
		}

		listResult := new(ListSprintResult)
		res, err := client.Do(req, listResult)
		body, _ := ioutil.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_sprint.listSprints", "res_body", string(body))
		if err != nil {
			if isNotFoundError(err) || strings.Contains(err.Error(), "400") {
				return nil, nil
			}
			plugin.Logger(ctx).Error("jira_sprint.listSprints", "api_error", err)
			return nil, err
		}

		for _, sprint := range listResult.Values {
			d.StreamListItem(ctx, SprintItemInfo{int64(board.ID), sprint})
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
