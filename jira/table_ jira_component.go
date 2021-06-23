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

func tableComponent(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_component",
		Description: "This resource represents project components. Use it to get, create, update, and delete project components. Also get components for project and get a count of issues by component.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getComponent,
		},
		List: &plugin.ListConfig{
			ParentHydrate: listProjects,
			Hydrate:       listComponents,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier for the component.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The name for the component.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description for the component.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL for this count of the issues contained in the component.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project",
				Description: "The key of the project to which the component is assigned.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "assignee_type",
				Description: "The nominal user type used to determine the assignee for issues created with this component.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_assignee_type_valid",
				Description: "Whether a user is associated with assigneeType.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "issue_count",
				Description: "The count of issues for the component.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "The ID of the project the component belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "real_assignee_type",
				Description: "The type of the assignee that is assigned to issues created with this component, when an assignee cannot be set from the assigneeType.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "assignee",
				Description: "The details of the user associated with assigneeType, if any.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "lead",
				Description: "The user details for the component's lead user.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "real_assignee",
				Description: "The user assigned to issues created with this component, when assigneeType does not identify a valid assignee.",
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

func listComponents(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("listComponents")

	project := h.Item.(Project)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {
		apiEndpoint := fmt.Sprintf("/rest/api/3/project/%s/component?startAt=%d&maxResults=%d", project.ID, last, maxResults)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return nil, err
		}

		listResult := new(ListComponentResult)
		_, err = client.Do(req, listResult)
		if err != nil {
			if isNotFoundError(err) {
				return nil, nil
			}
			logger.Error("listComponents", "Error", err)
			return nil, err
		}

		for _, component := range listResult.Values {
			d.StreamListItem(ctx, component)
		}

		last = listResult.StartAt + len(listResult.Values)
		if listResult.IsLast {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTIONS

func getComponent(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getComponent")
	componentId := d.KeyColumnQuals["id"]

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf(
		"rest/api/3/component/%s",
		componentId,
	)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(jira.Component)

	_, err = client.Do(req, result)
	if err != nil {
		plugin.Logger(ctx).Error("getComponent", "Error", err)
		return nil, err
	}

	return result, nil
}

type ListComponentResult struct {
	Self       string      `json:"self"`
	NextPage   string      `json:"nextPage"`
	MaxResults int         `json:"maxResults"`
	StartAt    int         `json:"startAt"`
	Total      int         `json:"total"`
	IsLast     bool        `json:"isLast"`
	Values     []Component `json:"values"`
}

type Component struct {
	IssueCount          int64     `json:"issueCount"`
	RealAssignee        jira.User `json:"realAssignee"`
	IsAssigneeTypeValid bool      `json:"isAssigneeTypeValid"`
	RealAssigneeType    string    `json:"realAssigneeType"`
	Description         string    `json:"description"`
	Project             string    `json:"project"`
	Self                string    `json:"self"`
	AssigneeType        string    `json:"assigneeType"`
	Lead                jira.User `json:"lead"`
	Assignee            jira.User `json:"assignee"`
	ProjectId           int64     `json:"projectId"`
	Name                string    `json:"name"`
	Id                  string    `json:"id"`
}
