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
		Name:        "jira_board",
		Description: "A board displays issues from one or more projects, giving you a flexible way of viewing, managing, and reporting on work in progress.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getBoard,
		},
		List: &plugin.ListConfig{
			Hydrate: listBoards,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The ID of the board.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The name of the board.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the board details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The board type of the board. Valid values are simple, scrum and kanban.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "filter_id",
				Description: "Filter id of the board.",
				Type:        proto.ColumnType_INT,
				Hydrate:     getBoardConfiguration,
				Transform:   transform.FromField("Filter.ID"),
			},
			{
				Name:        "sub_query",
				Description: "JQL subquery used by the given board - (Kanban only).",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getBoardConfiguration,
				Transform:   transform.FromField("SubQuery.Query"),
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
	logger := plugin.Logger(ctx)
	logger.Trace("listBoards")
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
			logger.Error("listBoards", "Error", err)
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

//// HYDRATE FUNCTIONS

func getBoard(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getBoard")
	boardId := d.KeyColumnQuals["id"].GetInt64Value()
	if boardId == 0 {
		return nil, nil
	}
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	board, _, err := client.Board.GetBoard(int(boardId))
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		logger.Error("getBoard", "Error", err)
		return nil, err
	}

	return *board, err
}

func getBoardConfiguration(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getBoardConfiguration")
	board := h.Item.(jira.Board)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	boardConfiguration, _, err := client.Board.GetBoardConfiguration(board.ID)
	if err != nil {
		logger.Error("getBoardConfiguration", "Error", err)
		return nil, err
	}

	return boardConfiguration, err
}
