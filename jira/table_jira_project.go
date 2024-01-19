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

func tableProject(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_project",
		Description: "Project is a collection of issues (stories, bugs, tasks, etc).",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getProject,
		},
		List: &plugin.ListConfig{
			Hydrate: listProjects,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "key", Require: plugin.Optional},
				{Name: "project_type_key", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The ID of the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The name of the project.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "key",
				Description: "The key of the project.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the project details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "A brief description of the project.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProject,
				Transform:   transform.FromCamel().NullIfZero(),
			},
			{
				Name:        "email",
				Description: "An email address associated with the project.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProject,
				Transform:   transform.FromCamel().NullIfZero(),
			},
			{
				Name:        "lead_account_id",
				Description: "The user account id of the project lead.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProject,
				Transform:   transform.FromField("Lead.AccountID"),
			},
			{
				Name:        "lead_display_name",
				Description: "The user display name of the project lead.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProject,
				Transform:   transform.FromField("Lead.DisplayName"),
			},
			{
				Name:        "project_type_key",
				Description: "The project type of the project. Valid values are software, service_desk and business.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "A link to information about this project, such as project documentation.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProject,
				Transform:   transform.FromField("url"),
			},

			// json fields
			{
				Name:        "component_ids",
				Description: "List of the components contained in the project.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getProject,
				Transform:   transform.FromField("Components").Transform(extractProjectComponentIds),
			},
			{
				Name:        "properties",
				Description: "This resource represents project properties, which provide for storing custom data against a project.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getProjectProperties,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "issue_types",
				Description: "List of the issue types available in the project.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "project_category",
				Description: "The category the project belongs to.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromCamel().NullIfZero(),
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

func listProjects(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_project.listProjects", "connection_error", err)
		return nil, err
	}

	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	queryLimit := d.QueryContext.Limit
	var maxResults int = 1000
	if d.QueryContext.Limit != nil {
		if *queryLimit < 1000 {
			maxResults = int(*queryLimit)
		}
	}

	query := ""
	if d.EqualsQualString("key") != "" {
		query = fmt.Sprintf("&%skeys=%s", query, d.EqualsQualString("key"))
	}
	if d.EqualsQualString("project_type_key") != "" {
		query = fmt.Sprintf("&%stypeKey=%s", query, d.EqualsQualString("project_type_key"))
	}

	projectCount := 1
	projectLimit := 200
	last := 0
	for {
		apiEndpoint := fmt.Sprintf(
			"rest/api/3/project/search?expand=description,lead,issueTypes,url,projectKeys,permissions,insight&startAt=%d&maxResults=%d", last,
			maxResults,
		)

		if query != "" {
			apiEndpoint = fmt.Sprintf("%s%s", apiEndpoint, query)
		}

		req, err := client.NewRequestWithContext(ctx, "GET", apiEndpoint, nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_project.listProjects", "get_request_error", err)
			return nil, err
		}

		projectList := new(ProjectListResult)
		res, err := client.Do(req, projectList)
		body, _ := io.ReadAll(res.Body)
		plugin.Logger(ctx).Debug("jira_project.listProjects", "res_body", string(body))

		if err != nil {
			plugin.Logger(ctx).Error("jira_project.listProjects", "api_error", err)
			return nil, err
		}

		for _, project := range projectList.Values {
			if projectCount > projectLimit {
				plugin.Logger(ctx).Debug("Maximum number of projects reached", projectLimit)
				return nil, nil
			}
			d.StreamListItem(ctx, project)
			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
			projectCount += 1
		}
		last = projectList.StartAt + len(projectList.Values)
		if projectList.IsLast || projectCount >= projectLimit {
			return nil, nil
		}
	}

}

//// HYDRATE FUNCTION

func getProject(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var projectId string
	if h.Item != nil {
		projectId = h.Item.(Project).ID
	} else {
		projectId = d.EqualsQuals["id"].GetStringValue()
	}

	if projectId == "" {
		return nil, nil
	}

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_project.getProject", "connection_error", err)
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/2/project/%s", projectId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_project.getProject", "get_request_error", err)
		return nil, err
	}

	project := new(Project)
	res, err := client.Do(req, project)
	body, _ := io.ReadAll(res.Body)
	plugin.Logger(ctx).Debug("jira_project.getProject", "res_body", string(body))

	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("jira_project.getProject", "api_error", err)
		return nil, err
	}

	return project, err
}

func getProjectProperties(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	project := getProjectInfo(ctx, h.Item)

	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("jira_project.getProjectProperties", "connection_error", err)
		return nil, err
	}

	keys, err := getProjectPropertyKeys(ctx, client, project.ID)
	if err != nil {
		return nil, err
	}

	var properties []KeyPropertyValue
	for _, key := range keys {
		apiEndpoint := fmt.Sprintf("rest/api/3/project/%s/properties/%s", project.ID, key.Key)

		req, err := client.NewRequest("GET", strings.Trim(apiEndpoint, " "), nil)
		if err != nil {
			plugin.Logger(ctx).Error("jira_project.getProjectProperties", "get_request_error", err)
			return nil, err
		}

		property := new(KeyPropertyValue)
		_, err = client.Do(req, property)
		if err != nil {
			plugin.Logger(ctx).Error("jira_project.getProjectProperties", "api_error", err)
			return nil, err
		}
		properties = append(properties, *property)

	}

	return properties, nil
}

func getProjectPropertyKeys(ctx context.Context, client *jira.Client, projectId string) ([]ProjectKey, error) {
	apiEndpoint := fmt.Sprintf("rest/api/3/project/%s/properties", projectId)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		plugin.Logger(ctx).Error("jira_project.getProjectPropertyKeys", "get_request_error", err)
		return nil, err
	}

	keys := new(ProjectKeys)
	_, err = client.Do(req, keys)
	if err != nil {
		plugin.Logger(ctx).Error("jira_project.getProjectPropertyKeys", "api_error", err)
		return nil, err
	}

	return keys.Keys, nil
}

//// TRANSFORM FUNCTION

func getProjectInfo(ctx context.Context, projectInfo interface{}) Project {
	switch item := projectInfo.(type) {
	case *Project:
		return *item
	case Project:
		return item
	}
	return Project{}
}

func extractProjectComponentIds(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var componentIds []string
	for _, item := range d.Value.([]jira.ProjectComponent) {
		componentIds = append(componentIds, item.ID)
	}
	return componentIds, nil
}

//// Custom Structs

// type ProjectListResult []Project
type ProjectListResult struct {
	MaxResults int       `json:"maxResults"`
	StartAt    int       `json:"startAt"`
	Total      int       `json:"total"`
	IsLast     bool      `json:"isLast"`
	Values     []Project `json:"values"`
}

// Project represents a Jira Project.
type Project struct {
	Expand          string                  `json:"expand,omitempty" structs:"expand,omitempty"`
	Self            string                  `json:"self,omitempty" structs:"self,omitempty"`
	ID              string                  `json:"id,omitempty" structs:"id,omitempty"`
	Key             string                  `json:"key,omitempty" structs:"key,omitempty"`
	Description     string                  `json:"description,omitempty" structs:"description,omitempty"`
	Lead            jira.User               `json:"lead,omitempty" structs:"lead,omitempty"`
	Components      []jira.ProjectComponent `json:"components,omitempty" structs:"components,omitempty"`
	IssueTypes      []jira.IssueType        `json:"issueTypes,omitempty" structs:"issueTypes,omitempty"`
	URL             string                  `json:"url,omitempty" structs:"url,omitempty"`
	Email           string                  `json:"email,omitempty" structs:"email,omitempty"`
	AssigneeType    string                  `json:"assigneeType,omitempty" structs:"assigneeType,omitempty"`
	Versions        []jira.Version          `json:"versions,omitempty" structs:"versions,omitempty"`
	Name            string                  `json:"name,omitempty" structs:"name,omitempty"`
	Roles           map[string]string       `json:"roles,omitempty" structs:"roles,omitempty"`
	AvatarUrls      jira.AvatarUrls         `json:"avatarUrls,omitempty" structs:"avatarUrls,omitempty"`
	ProjectCategory jira.ProjectCategory    `json:"projectCategory,omitempty" structs:"projectCategory,omitempty"`
	ProjectTypeKey  string                  `json:"projectTypeKey" structs:"projectTypeKey"`
}

type KeyPropertyValue struct {
	Key   string      `json:"key,omitempty" structs:"key,omitempty"`
	Value interface{} `json:"value,omitempty" structs:"value,omitempty"`
}

type ProjectKeys struct {
	Keys []ProjectKey `json:"keys,omitempty" structs:"keys,omitempty"`
}
type ProjectKey struct {
	Self string `json:"self,omitempty" structs:"self,omitempty"`
	Key  string `json:"key,omitempty" structs:"key,omitempty"`
}
