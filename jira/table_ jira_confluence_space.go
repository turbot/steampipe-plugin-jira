package jira

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableConfluenceSpace(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_confluence_space",
		Description: "Spaces are collections of related pages that you and other people in your team or organization work on together.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("key"),
			Hydrate:    getSpace,
		},
		List: &plugin.ListConfig{
			Hydrate: listSpaces,
		},
		Columns: []*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the space.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "key",
				Description: "The key of the space.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The name of the space.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the space.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Description.plain.value"),
			},

			// other important fields
			{
				Name:        "self",
				Type:        proto.ColumnType_STRING,
				Description: "The URL of the confluence space details.",
				Transform:   transform.FromField("Links.self"),
			},
			{
				Name:        "status",
				Description: "The current status of the space.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The type of the space.",
				Type:        proto.ColumnType_STRING,
			},

			// JSON fields
			{
				Name:        "category",
				Description: "The category of the space.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(extractSpaceCategory),
			},
			{
				Name:        "homepage",
				Description: "The homepage details of the space.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(extractHomepageDetails),
			},
			{
				Name:        "icon",
				Description: "The details of the icon of the space.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "operations",
				Description: "The details of the operations of the space.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "permissions",
				Type:        proto.ColumnType_JSON,
				Description: "The details of the permissions of the space.",
			},
			{
				Name:        "settings",
				Type:        proto.ColumnType_JSON,
				Description: "The details of the settings of the space.",
			},

			// Standard columns
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

func listSpaces(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("listSpaces")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {

		apiEndpoint := fmt.Sprintf(
			"wiki/rest/api/space?start=%d&limit=%d&expand=settings,metadata.labels,operations,lookAndFeel,permissions,icon,description.plain,description.view,theme,homepage", last, maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			if isNotFoundError(err) {
				return nil, nil
			}
			logger.Error("listSpaces", "Error", err)
			return nil, err
		}

		listSpacesResult := new(ListSpacesResult)
		_, err = client.Do(req, listSpacesResult)
		if err != nil {
			return nil, err
		}

		for _, space := range listSpacesResult.Spaces {
			d.StreamListItem(ctx, space)
		}

		last = listSpacesResult.Start + len(listSpacesResult.Spaces)
		if listSpacesResult.Size < listSpacesResult.Limit {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getSpace(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getSpace")

	spaceKey := d.KeyColumnQuals["key"].GetStringValue()

	if spaceKey == "" {
		return nil, nil
	}
	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf(
		"wiki/rest/api/space/%s?expand=settings,metadata.labels,operations,lookAndFeel,permissions,icon,description.plain,theme,homepage", spaceKey,
	)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		logger.Error("getSpace", "Error", err)
		return nil, err
	}

	spacesResult := new(Space)
	_, err = client.Do(req, spacesResult)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return spacesResult, err
}

//// TRANSFORM FUNCTION

func extractSpaceCategory(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var category []string
	category = []string{}
	spaceMetadata := d.HydrateItem.(Space).Metadata
	for _, item := range spaceMetadata.Labels.Results {
		category = append(category, item["label"].(string))
	}
	return category, nil
}

func extractHomepageDetails(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var homePageDetails HomepageDetails
	homePageData := d.HydrateItem.(Space).Homepage

	homePageDetails.ID = homePageData.ID
	homePageDetails.Type = homePageData.Type
	homePageDetails.Title = homePageData.Title
	homePageDetails.Status = homePageData.Status
	homePageDetails.Self = homePageData.Links.Self

	return homePageDetails, nil
}

//// Required Structs

type ListSpacesResult struct {
	Start  int     `json:"start"`
	Limit  int     `json:"limit"`
	Size   int     `json:"size"`
	Spaces []Space `json:"results"`
}

type Space struct {
	ID          int                    `json:"id,omitempty"`
	Key         string                 `json:"key,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Icon        map[string]interface{} `json:"icon,omitempty"`
	Description map[string]interface{} `json:"description,omitempty"`
	Homepage    Homepage               `json:"homepage,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Metadata    Metadata               `json:"metadata,omitempty"`
	Operations  interface{}            `json:"operations,omitempty"`
	Permissions interface{}            `json:"permissions,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	LookAndFeel map[string]interface{} `json:"lookAndFeel,omitempty"`
	Links       map[string]interface{} `json:"_links,omitempty"`
}

type Homepage struct {
	ID     string        `json:"id"`
	Type   string        `json:"type"`
	Title  string        `json:"title"`
	Status string        `json:"status"`
	Links  HomepageLinks `json:"_links"`
}

type HomepageLinks struct {
	Self string `json:"self"`
}

type HomepageDetails struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Self   string `json:"self"`
}

type Metadata struct {
	Labels MetadataLabels `json:"labels,omitempty"`
}

type MetadataLabels struct {
	Results []map[string]interface{} `json:"results,omitempty"`
}
