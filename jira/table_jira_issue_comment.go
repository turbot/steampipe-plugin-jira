package jira

import (
	"context"
	"fmt"
	"io"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableIssueComment(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_issue_comment",
		Description: "It helps to get, create, update, and delete a comment from an issue.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getIssueComment,
		},
		List: &plugin.ListConfig{
			Hydrate:       listIssueComments,
			ParentHydrate: listIssues,
		},
		Columns: []*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the comment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the comment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the comment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "author",
				Description: "The ID of the user who created the comment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "body",
				Description: "The comment text in Atlassian Document Format.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "update_author",
				Description: "The ID of the user who updated the comment last.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created",
				Description: "The date and time at which the comment was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Created").NullIfZero().Transform(convertJiraDate),
			},
			{
				Name:        "updated",
				Description: "The update time of the issue comment.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Updated").NullIfZero().Transform(convertJiraDate),
			},
			{
				Name:        "visibility",
				Description: "The visibility of the issue comment.",
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

func listIssueComments(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	issueId := h.Item.(IssueInfo).Issue.ID

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_comment.listIssueComments", "connection_error", err)
		return nil, err
	}

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

	for {
		apiEndpoint := fmt.Sprintf("/rest/api/3/issue/%s/comment?startAt=%d&maxResults=%d", issueId, last, maxResults)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_comment.listIssueComments", "get_request_error", err)
			return nil, err
		}

		listResult := new(ListIssuesCommentResult)
		res, err := client.Do(req, listResult)
		body, _ := io.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_issue_comment.listIssueComments", "res_body", string(body))

		if err != nil {
			plugin.Logger(ctx).Error("jira_issue_comment.listIssueComments", "api_error", err)
			return nil, err
		}

		for _, comment := range listResult.Comments {
			d.StreamListItem(ctx, comment)
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		last = listResult.StartAt + len(listResult.Comments)
		if last >= listResult.Total {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getIssueComment(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var issueId, issueCommentID string
	issueId = h.Item.(IssueInfo).Issue.ID
	issueCommentID = d.EqualsQualString("id")

	if issueCommentID == "" {
		return nil, nil
	}
	issueComment := new(ListIssuesCommentResult)
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_comment.getIssueComment", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/issue/%s/comment/%s", issueId, issueCommentID)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_issue_comment.getIssueComment", "get_request_error", err)
		return nil, err
	}

	res, err := client.Do(req, issueComment)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_issue_comment.getIssueComment", "res_body", string(body))

	if err != nil && isNotFoundError(err) {
		plugin.Logger(ctx).Error("jira_issue_comment.getIssueComment", "api_error", err)
		return nil, nil
	}

	return issueComment, err
}

//// Required Structs

type ListIssuesCommentResult struct {
	Expand     string         `json:"expand"`
	MaxResults int            `json:"maxResults"`
	StartAt    int            `json:"startAt"`
	Total      int            `json:"total"`
	Comments   []jira.Comment `json:"comments"`
}

type CommentInfo struct {
	jira.Comment
	Keys map[string]string
}
