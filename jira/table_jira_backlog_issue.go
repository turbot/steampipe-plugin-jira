package jira

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v4/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
)

//// TABLE DEFINITION

func tableBacklogIssue(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_backlog_issue",
		Description: "The backlog contains incomplete issues that are not assigned to any future or active sprint.",
		List: &plugin.ListConfig{
			ParentHydrate: listBoards,
			Hydrate:       listBacklogIssues,
		},
		Columns: []*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "key",
				Description: "The key of the issue.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the issue details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "board_name",
				Description: "The name of the board the issue belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("BoardName"),
			},
			{
				Name:        "board_id",
				Description: "The ID of the board the issue belongs to.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("BoardId"),
			},
			{
				Name:        "project_key",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Key"),
			},
			{
				Name:        "project_id",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.ID"),
			},
			{
				Name:        "status",
				Description: "The status of the issue. Eg: To Do, In Progress, Done.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Status.Name"),
			},

			// other important fields
			{
				Name:        "assignee_account_id",
				Description: "Account Id the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Assignee.AccountID"),
			},
			{
				Name:        "assignee_display_name",
				Description: "Display name the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Assignee.DisplayName"),
			},
			{
				Name:        "created",
				Description: "Time when the issue was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Created").Transform(convertJiraTime),
			},
			{
				Name:        "creator_account_id",
				Description: "Account Id of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.AccountID"),
			},
			{
				Name:        "creator_display_name",
				Description: "Display name of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.DisplayName"),
			},
			{
				Name:        "description",
				Description: "Description of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Description"),
			},
			{
				Name:        "due_date",
				Description: "Time by which the issue is expected to be completed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Duedate").NullIfZero().Transform(convertJiraDate),
			},
			{
				Name:        "epic_key",
				Description: "The key of the epic to which issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromP(extractBacklogIssueRequiredField, "epic"),
			},
			{
				Name:        "priority",
				Description: "Priority assigned to the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Priority.Name"),
			},
			{
				Name:        "project_name",
				Description: "Name of the project to that issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Name"),
			},
			{
				Name:        "reporter_account_id",
				Description: "Account Id of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.AccountID"),
			},
			{
				Name:        "reporter_display_name",
				Description: "Display name of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.DisplayName"),
			},
			{
				Name:        "summary",
				Description: "Details of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Summary"),
			},
			{
				Name:        "type",
				Description: "The name of the issue type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Type.Name"),
			},
			{
				Name:        "updated",
				Description: "Time when the issue was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Updated").Transform(convertJiraTime),
			},

			// JSON fields
			{
				Name:        "components",
				Description: "List of components associated with the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Components").Transform(extractComponentIds),
			},
			{
				Name:        "fields",
				Description: "Json object containing important subfields of the issue.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "labels",
				Description: "A list of labels applied to the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Labels"),
			},

			// Standard columns
			{
				Name:        "tags",
				Type:        proto.ColumnType_JSON,
				Description: "A map of label names associated with this issue, in Steampipe standard format.",
				Transform:   transform.From(getBacklogIssueTags),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Key"),
			},
		},
	}
}

//// LIST FUNCTION

func listBacklogIssues(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_backlog_issue.listBacklogIssues", "connection_error", err)
		return nil, err
	}

	board := h.Item.(jira.Board)

	last := 0

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var maxResults int = 1000
	if d.QueryContext.Limit != nil {
		if *queryLimit < 1000 {
			maxResults = int(*queryLimit)
		}
	}

	var epicKey string
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/agile/1.0/board/%d/backlog?startAt=%d&maxResults=%d&expand=names",
			board.ID,
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			if isNotFoundError(err) || strings.Contains(err.Error(), "400") {
				return nil, nil
			}
			plugin.Logger(ctx).Error("jira_backlog_issue.listBacklogIssues", "get_request_error", err)
			return nil, err
		}

		listIssuesResult := new(ListIssuesResult)
		res, err := client.Do(req, listIssuesResult)
		body, _ := io.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_backlog_issue.listBacklogIssues", "res_body", string(body))

		if err != nil {
			plugin.Logger(ctx).Error("jira_backlog_issue.listBacklogIssues", "api_error", err)
			return nil, err
		}

		epicKey = getFieldKey(ctx, d, listIssuesResult.Names, "Epic Link")

		keys := map[string]string{
			"epic": epicKey,
		}

		for _, issue := range listIssuesResult.Issues {
			d.StreamListItem(ctx, BacklogIssueInfo{issue, board.ID, board.Name, keys})
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.QueryStatus.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		last = listIssuesResult.StartAt + len(listIssuesResult.Issues)
		if last >= listIssuesResult.Total {
			return nil, nil
		}
	}
}

//// TRANSFORM FUNCTION

func getBacklogIssueTags(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issue := d.HydrateItem.(BacklogIssueInfo)

	tags := make(map[string]bool)
	if issue.Fields != nil && issue.Fields.Labels != nil {
		for _, i := range issue.Fields.Labels {
			tags[i] = true
		}
	}
	return tags, nil
}

func extractBacklogIssueRequiredField(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issueInfo := d.HydrateItem.(BacklogIssueInfo)
	m := issueInfo.Fields.Unknowns
	param := d.Param.(string)
	return m[issueInfo.Keys[param]], nil
}

//// Required Structs

type BacklogIssueInfo struct {
	jira.Issue
	BoardId   int
	BoardName string
	Keys      map[string]string
}
