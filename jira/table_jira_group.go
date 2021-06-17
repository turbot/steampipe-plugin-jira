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
	logger := plugin.Logger(ctx)
	logger.Trace("listGroups")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/api/3/group/bulk?startAt=%d&maxResults=%d",
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return nil, err
		}

		listGroupResult := new(ListGroupResult)
		_, err = client.Do(req, listGroupResult)
		if err != nil {
			logger.Error("listGroups", "Error", err)
			return nil, err
		}

		for _, group := range listGroupResult.Groups {
			d.StreamListItem(ctx, group)
		}

		last = listGroupResult.StartAt + len(listGroupResult.Groups)
		if last >= listGroupResult.Total {
			return nil, nil
		}
	}
}

//// HDRATE FUNCTIONS

func getGroup(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getGroup")

	groupId := d.KeyColumnQuals["id"].GetStringValue()

	if groupId == "" {
		return nil, nil
	}

	listGroupResult := new(ListGroupResult)
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}
	apiEndpoint := fmt.Sprintf("/rest/api/3/group/bulk?groupId=%s", groupId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		logger.Error("getGroup", "Error", err)
		return nil, err
	}

	_, err = client.Do(req, listGroupResult)
	if err != nil {
		return nil, err
	}

	if len(listGroupResult.Groups) > 0 {
		return listGroupResult.Groups[0], nil
	}
	return nil, nil
}

func getGroupMembers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getGroupMembers")
	group := h.Item.(Group)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	groupMembers := []jira.GroupMember{}

	last := 0
	for {
		opts := &jira.GroupSearchOptions{
			MaxResults:           1000,
			StartAt:              last,
			IncludeInactiveUsers: true,
		}

		chunk, resp, err := client.Group.GetWithOptions(group.Name, opts)
		if err != nil {
			if isNotFoundError(err) {
				return groupMembers, nil
			}
			logger.Error("getGroupMembers", "Error", err)
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
