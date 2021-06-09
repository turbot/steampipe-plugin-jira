package jira

import (
	"context"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableBoard(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_board",
		Description:      "Jira Board",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getBoard,
		},
		List: &plugin.ListConfig{
			Hydrate: listBoards,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "A friendly name that identifies the board.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "id",
				Description: "The unique identifier of board.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromGo(),
			},
			{
				Name: "self",
				Type: proto.ColumnType_STRING,
			},
			{
				Name: "type",
				Type: proto.ColumnType_STRING,
			},
			{
				Name:      "filter_id",
				Type:      proto.ColumnType_INT,
				Transform: transform.FromGo(),
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

func listBoards(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	for {
		opt := jira.SearchOptions{
			MaxResults: 1000,
			StartAt:    last,
		}

		boardList, resp, err := client.Board.GetAllBoardsWithContext(ctx, &jira.BoardListOptions{
			SearchOptions: opt,
		})
		if err != nil {
			return nil, err
		}

		total := resp.Total

		for _, board := range boardList.Values {
			d.StreamListItem(ctx, board)
		}

		last = resp.StartAt + len(boardList.Values)
		if last >= total {
			return nil, nil
		}
	}
}

func getBoard(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	boardId := d.KeyColumnQuals["id"].GetInt64Value()
	if boardId == 0 {
		return nil, nil
	}
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	board, _, err := client.Board.GetBoard(int(boardId))
	if err != nil && isNotFoundError(err) {
		return nil, nil
	}

	return board, err
}
