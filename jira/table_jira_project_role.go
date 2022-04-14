package jira

import (
	"context"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v3/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v3/plugin"
)

//// TABLE DEFINITION

func tableProjectRole(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_project_role",
		Description: "Project Roles are a flexible way to associate users and/or groups with particular projects.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getProjectRole,
		},
		List: &plugin.ListConfig{
			Hydrate: listProjectRoles,
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The ID of the project role.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The name of the project role.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL the project role details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the project role.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "actor_account_ids",
				Description: "The list of user ids who act in this role.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(extractActorAccountIds),
			},
			{
				Name:        "actor_names",
				Description: "The list of user ids who act in this role.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(extractActorNames),
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

func listProjectRoles(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("listProjectRoles")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	roles, _, err := client.Role.GetList()
	if err != nil {
		logger.Error("listProjectRoles", "Error", err)
		return nil, err
	}

	for _, role := range *roles {
		d.StreamListItem(ctx, role)
	}

	return nil, err
}

//// HYDRATE FUNCTION

func getProjectRole(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getProjectRole")

	roleId := d.KeyColumnQuals["id"].GetInt64Value()

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	if roleId == 0 {
		return nil, nil
	}

	role, _, err := client.Role.Get(int(roleId))
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		logger.Error("getProjectRole", "Error", err)
		return nil, err
	}

	return *role, err
}

//// TRANSFORM FUNCTION

func extractActorAccountIds(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var actorIds []string
	for _, actor := range d.HydrateItem.(jira.Role).Actors {
		actorIds = append(actorIds, actor.ActorUser.AccountID)
	}
	return actorIds, nil
}

func extractActorNames(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var actorNames []string
	for _, actor := range d.HydrateItem.(jira.Role).Actors {
		actorNames = append(actorNames, actor.DisplayName)
	}
	return actorNames, nil
}
