package jira

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
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
			// top fields
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

			// other important fields
			{
				Name:        "assignee_account_id",
				Description: "The account id of the user associated with assigneeType, if any.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Assignee.AccountID"),
			},
			{
				Name:        "assignee_display_name",
				Description: "The display name of the user associated with assigneeType, if any.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Assignee.DisplayName"),
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
				Name:        "lead_account_id",
				Description: "The account id for the component's lead user.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Lead.AccountID"),
			},
			{
				Name:        "lead_display_name",
				Description: "The display name for the component's lead user.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Lead.DisplayName"),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project the component belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "real_assignee_account_id",
				Description: "The account id of the user assigned to issues created with this component, when assigneeType does not identify a valid assignee.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RealAssignee.AccountID"),
			},
			{
				Name:        "real_assignee_display_name",
				Description: "The display name of the user assigned to issues created with this component, when assigneeType does not identify a valid assignee.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RealAssignee.DisplayName"),
			},
			{
				Name:        "real_assignee_type",
				Description: "The type of the assignee that is assigned to issues created with this component, when an assignee cannot be set from the assigneeType.",
				Type:        proto.ColumnType_STRING,
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
	project := h.Item.(Project)

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_component.listComponents", "connection_error", err)
		return nil, err
	}

	componentCount := 1
	componentLimit, err := getComponentLimit(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_component.listComponents", "component_limit", err)
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
		apiEndpoint := fmt.Sprintf("/rest/api/3/project/%s/component?startAt=%d&maxResults=%d", project.ID, last, maxResults)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_component.listComponents", "get_request_error", err)
			return nil, err
		}

		listResult := new(ListComponentResult)
		res, err := client.Do(req, listResult)
		body, _ := io.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_component.listComponents", "res_body", string(body))

		if err != nil {
			if isNotFoundError(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("jira_component.listComponents", "api_error", err)
			return nil, err
		}

		// return error if user requests too much data
		if listResult.Total > componentLimit {
			return nil, fmt.Errorf("Number of results exceeds component limit(%d>%d). Please make your query more specific.", listResult.Total, componentLimit)
		}

		sensitivity, err := getCaseSensitivity(ctx, d)
		if err != nil {
			return nil, err
		}
		plugin.Logger(ctx).Debug("jira_component.listComponents", "case_sensitivity", sensitivity)

		for _, component := range listResult.Values {
			if componentCount > componentLimit {
				plugin.Logger(ctx).Debug("Maximum number of components reached", componentLimit)
				return nil, nil
			}
			if sensitivity == "insensitive" {
				component.RealAssigneeType = strings.ToLower(component.RealAssigneeType)
				component.Description = strings.ToLower(component.Description)
				component.Project = strings.ToLower(component.Project)
				component.Self = strings.ToLower(component.Self)
				component.AssigneeType = strings.ToLower(component.AssigneeType)
				component.Name = strings.ToLower(component.Name)
				component.Id = strings.ToLower(component.Id)
				component.Assignee.AccountID = strings.ToLower(component.Assignee.AccountID)
				component.Assignee.DisplayName = strings.ToLower(component.Assignee.DisplayName)
				component.Lead.AccountID = strings.ToLower(component.Lead.AccountID)
				component.Lead.DisplayName = strings.ToLower(component.Lead.DisplayName)
				component.RealAssignee.DisplayName = strings.ToLower(component.RealAssignee.DisplayName)
				component.RealAssignee.AccountID = strings.ToLower(component.RealAssignee.AccountID)
			}
			d.StreamListItem(ctx, component)

			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
			componentCount += 1
		}

		last = listResult.StartAt + len(listResult.Values)
		if listResult.IsLast {
			return nil, nil
		} else if componentCount >= componentLimit {
			plugin.Logger(ctx).Debug("Maximum number of components reached", componentLimit)
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTIONS

func getComponent(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	componentId := d.EqualsQuals["id"].GetStringValue()

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_component.getComponent", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/component/%s", componentId)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_component.getComponent", "get_request_error", err)
		return nil, err
	}

	result := new(Component)

	res, err := client.Do(req, result)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_component.getComponent", "res_body", string(body))

	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_component.getComponent", "api_error", err)
		return nil, err
	}

	sensitivity, err := getCaseSensitivity(ctx, d)
	if err != nil {
		return nil, err
	}
	plugin.Logger(ctx).Debug("jira_component.getComponent", "case_sensitivity", sensitivity)

	if sensitivity == "insensitive" {
		result.RealAssigneeType = strings.ToLower(result.RealAssigneeType)
		result.Description = strings.ToLower(result.Description)
		result.Project = strings.ToLower(result.Project)
		result.Self = strings.ToLower(result.Self)
		result.AssigneeType = strings.ToLower(result.AssigneeType)
		result.Name = strings.ToLower(result.Name)
		result.Id = strings.ToLower(result.Id)
		result.Assignee.AccountID = strings.ToLower(result.Assignee.AccountID)
		result.Assignee.DisplayName = strings.ToLower(result.Assignee.DisplayName)
		result.Lead.AccountID = strings.ToLower(result.Lead.AccountID)
		result.Lead.DisplayName = strings.ToLower(result.Lead.DisplayName)
		result.RealAssignee.DisplayName = strings.ToLower(result.RealAssignee.DisplayName)
		result.RealAssignee.AccountID = strings.ToLower(result.RealAssignee.AccountID)
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
