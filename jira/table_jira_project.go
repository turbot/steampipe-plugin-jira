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

func tableProject(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_project",
		Description:      "Jira Project",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getProject,
		},
		List: &plugin.ListConfig{
			Hydrate: listProjects,
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
				Name:        "project_type_key",
				Description: "The project type of the project. Valid values are software, service_desk and business.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "A link to information about this project, such as project documentation.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProject,
			},

			// json fields
			{
				Name:        "avatar_urls",
				Description: "The URLs of the project's avatars.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "components",
				Description: "List of the components contained in the project.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getProject,
			},
			{
				Name:        "issue_types",
				Description: "List of the issue types available in the project.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "lead",
				Description: "The user details of the project lead.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getProject,
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
		return nil, err
	}

	req, err := client.NewRequest("GET", "/rest/api/3/project", nil)
	if err != nil {
		return nil, err
	}

	projectList := new(ProjectList)
	_, err = client.Do(req, projectList)
	if err != nil {
		return nil, err
	}

	for _, project := range *projectList {
		d.StreamListItem(ctx, project)
	}

	return nil, nil
}

//// HYDRATE FUNCTION

func getProject(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var projectId string
	if h.Item != nil {
		projectId = h.Item.(Project).ID
	} else {
		projectId = d.KeyColumnQuals["id"].GetStringValue()
	}

	if projectId == "" {
		return nil, nil
	}

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("/rest/api/3/project/%s", projectId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	project := new(Project)
	_, err = client.Do(req, project)
	if err != nil && isNotFoundError(err) {
		return nil, nil
	}

	return project, err
}

type ProjectList []Project

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
