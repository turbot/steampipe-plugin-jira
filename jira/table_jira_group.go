package jira

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v4/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
)

//// TABLE DEFINITION

func tableGroup(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_group",
		Description: "Group is a collection of users. Administrators create groups so that the administrator can assign permissions to a number of people at once.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getGroup,
		},
		List: &plugin.ListConfig{
			Hydrate: listGroups,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the group.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The ID of the group, if available, which uniquely identifies the group across all Atlassian products. For example, 952d12c3-5b5b-4d04-bb32-44d383afc4b2.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GroupId"),
			},
			{
				Name:        "member_ids",
				Description: "List of account ids of users associated with the group.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getGroupMembers,
				Transform:   transform.From(memberIds),
			},
			{
				Name:        "member_names",
				Description: "List of names of users associated with the group.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getGroupMembers,
				Transform:   transform.From(memberNames),
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

func listGroups(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_group.listGroups", "connection_error", err)
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
		apiEndpoint := fmt.Sprintf(
			"/rest/api/3/group/bulk?startAt=%d&maxResults=%d",
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_group.listGroups", "get_request_error", err)
			return nil, err
		}

		listGroupResult := new(ListGroupResult)
		res, err := client.Do(req, listGroupResult)
		body, _ := ioutil.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_group.listGroups", "res_body", string(body))

		if err != nil {
			plugin.Logger(ctx).Error("jira_group.listGroups", "api_error", err)
			return nil, err
		}

		for _, group := range listGroupResult.Groups {
			d.StreamListItem(ctx, group)
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.QueryStatus.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		last = listGroupResult.StartAt + len(listGroupResult.Groups)
		if last >= listGroupResult.Total {
			return nil, nil
		}
	}
}

//// HDRATE FUNCTIONS

func getGroup(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	groupId := d.KeyColumnQuals["id"].GetStringValue()

	if groupId == "" {
		return nil, nil
	}

	listGroupResult := new(ListGroupResult)
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_group.getGroup", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/group/bulk?groupId=%s", groupId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_group.getGroup", "get_request_error", err)
		return nil, err
	}

	res, err := client.Do(req, listGroupResult)
	body, _ := ioutil.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_group.getGroup", "res_body", string(body))

	if err != nil {
		plugin.Logger(ctx).Error("jira_group.getGroup", "api_error", err)
		return nil, err
	}

	if len(listGroupResult.Groups) > 0 {
		return listGroupResult.Groups[0], nil
	}
	return nil, nil
}

func getGroupMembers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	group := h.Item.(Group)

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_group.getGroupMembers", "connection_error", err)
		return nil, err
	}

	groupMembers := []jira.GroupMember{}

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
		opts := &jira.GroupSearchOptions{
			MaxResults:           maxResults,
			StartAt:              last,
			IncludeInactiveUsers: true,
		}

		chunk, resp, err := client.Group.GetWithOptions(group.Name, opts)
		if err != nil {
			if isNotFoundError(err) {
				return groupMembers, nil
			}
			plugin.Logger(ctx).Error("jira_group.getGroupMembers", "api_error", err)
			return nil, err
		}

		if groupMembers == nil {
			groupMembers = make([]jira.GroupMember, 0, resp.Total)
		}

		groupMembers = append(groupMembers, chunk...)

		last = resp.StartAt + len(chunk)
		if last >= resp.Total {
			return groupMembers, nil
		}
	}
}

//// TRANSFORM FUNCTION

func memberIds(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var memberIds []string
	for _, member := range d.HydrateItem.([]jira.GroupMember) {
		memberIds = append(memberIds, member.AccountID)
	}
	return memberIds, nil
}

func memberNames(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var memberNames []string
	for _, member := range d.HydrateItem.([]jira.GroupMember) {
		memberNames = append(memberNames, member.DisplayName)
	}
	return memberNames, nil
}

//// Required Structs

type ListGroupResult struct {
	MaxResults int     `json:"maxResults"`
	StartAt    int     `json:"startAt"`
	Total      int     `json:"total"`
	IsLast     bool    `json:"isLast"`
	Groups     []Group `json:"values"`
}

type Group struct {
	Name    string `json:"name"`
	GroupId string `json:"groupId"`
}
