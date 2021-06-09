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
		Name:             "jira_group",
		Description:      "Jira Group",
		DefaultTransform: transform.FromCamel(),
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
				Description: "Friendly name of the atlassian group.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "Unique identifier of the atlassian group.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GroupId"),
			},
			{
				Name:        "group_members",
				Description: "Members associated with the group.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getGroupMembers,
				Transform:   transform.FromValue(),
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
