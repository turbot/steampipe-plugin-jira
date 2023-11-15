package jira

import (
	"context"
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableIssueComment(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_issue_comment",
		Description: "Comments that provided in issue.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AnyColumn([]string{"issue_id", "id"}),
			Hydrate:    getIssueComment,
		},
		List: &plugin.ListConfig{
			ParentHydrate: listIssues,
			Hydrate:       listIssueComments,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "issue_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the issue comment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "issue_id",
				Description: "The ID of the issue.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the issue comment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "body",
				Description: "The content of the issue comment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created",
				Description: "Time when the issue comment was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated",
				Description: "Time when the issue comment was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "jsd_public",
				Description: "JsdPublic set to false does not hide comments in Service Desk projects.",
				Type:        proto.ColumnType_BOOL,
			},

			// JSON fields
			{
				Name:        "author",
				Description: "The user information who added the issue comment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "update_author",
				Description: "The user information who updated the issue comment.",
				Type:        proto.ColumnType_JSON,
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
		},
	}
}

type CommentResult struct {
	Comments   []Comment `json:"comments" structs:"comments"`
	StartAt    int       `json:"startAt" structs:"startAt"`
	MaxResults int       `json:"maxResults" structs:"maxResults"`
	Total      int       `json:"total" structs:"total"`
}

type Comment struct {
	ID           string
	Self         string
	Author       jira.User
	Body         string
	UpdateAuthor jira.User
	Updated      string
	Created      string
	JsdPublic    bool
}

type commentWithIssueDetails struct {
	Comment
	IssueId string
}

//// LIST FUNCTION

func listIssueComments(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if h.Item == nil {
		return nil, nil
	}
	issueinfo := h.Item.(IssueInfo)
	issueId := d.EqualsQualString("issue_id")

	// Minize the API call for given issue ID.
	if issueId != "" && issueId != issueinfo.ID {
		return nil, nil
	}

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_comment.listIssueComments", "connection_error", err)
		return nil, err
	}

	last := 0

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var limit int = 5000
	if d.QueryContext.Limit != nil {
		if *queryLimit < 5000 {
			limit = int(*queryLimit)
		}
	}

	for {
		apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/comment?startAt=%d&maxResults=%d", issueinfo.ID, last, limit)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_comment.listIssueComments", "get_request_error", err)
			return nil, err
		}

		comments := new(CommentResult)
		_, err = client.Do(req, comments)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_comment.listIssueComments", "api_error", err)
			return nil, err
		}

		for _, c := range comments.Comments {
			d.StreamListItem(ctx, commentWithIssueDetails{c, issueinfo.ID})

			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		last = comments.StartAt + len(comments.Comments)
		if last >= comments.Total {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getIssueComment(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	issueId := d.EqualsQualString("issue_id")
	id := d.EqualsQualString("id")

	if issueId == "" || id == "" {
		return nil, nil
	}

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_comment.getIssueComment", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/comment/%s", issueId, id)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_comment.getIssueComment", "get_request_error", err)
		return nil, err
	}

	res := new(Comment)
	_, err = client.Do(req, res)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_issue_comment.getIssueComment", "api_error", err)
		return nil, err
	}

	return commentWithIssueDetails{*res, issueId}, nil
}
