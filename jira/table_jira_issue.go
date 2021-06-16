package jira

import (
	"context"
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableIssue(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "jira_issue",
		Description:      "Issues help manage code, estimate workload, and keep track of team.",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getIssue,
		},
		List: &plugin.ListConfig{
			// KeyColumns: plugin.SingleColumn("project_key"),
			Hydrate: listIssues,
		},
		Columns: []*plugin.Column{
			// top fields
			{
				Name:        "id",
				Description: "The ID of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "key",
				Description: "The key of the issue.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "self",
				Description: "The URL of the issue details.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_key",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Key"),
			},
			{
				Name:        "project_id",
				Description: "A friendly key that identifies the project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.ID"),
			},
			{
				Name:        "status",
				Description: "Json object containing important subfields info the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Status.Name"),
			},
			{
				Name:        "epic_id",
				Description: "The id of the epic to which issue belongs.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Fields.Epic.ID", "EpicID"),
			},
			{
				Name:        "epic_key",
				Description: "The key of the epic to which issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Epic.Key", "EpicKey"),
			},
			{
				Name:        "sprint_id",
				Description: "The ID of the sprint to which issue belongs.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Fields.Sprint.ID", "SprintID"),
			},

			// other important fields
			{
				Name:        "assignee_account_id",
				Description: "Account Id the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Assignee.AccountID"),
			},
			{
				Name:        "assignee_display_name",
				Description: "Display name the user/application that the issue is assigned to work.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Assignee.DisplayName"),
			},
			{
				Name:        "creator_account_id",
				Description: "Account Id of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.AccountID"),
			},
			{
				Name:        "creator_display_name",
				Description: "Display name of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Creator.DisplayName"),
			},
			{
				Name:        "created",
				Description: "Time when the issue was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Created").Transform(convertJiraTime),
			},
			{
				Name:        "duedate",
				Description: "Time by which the issue is expected to be completed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Duedate").Transform(convertJiraDate),
			},
			{
				Name:        "description",
				Description: "Description of the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Description"),
			},
			{
				Name:        "type",
				Description: "The name of the issue type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Type.Name"),
			},
			{
				Name:        "labels",
				Description: "A list of labels applied to the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Labels"),
			},
			{
				Name:        "priority",
				Description: "Priority assigned to the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Priority.Name"),
			},
			{
				Name:        "project_name",
				Description: "Name of the project to that issue belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Project.Name"),
			},
			{
				Name:        "reporter_account_id",
				Description: "Account Id of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.AccountID"),
			},
			{
				Name:        "reporter_display_name",
				Description: "Display name of the user/application issue is reported.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Reporter.DisplayName"),
			},
			{
				Name:        "summary",
				Description: "Details of the user/application that created the issue.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Fields.Summary"),
			},
			{
				Name:        "updated",
				Description: "Time when the issue was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Fields.Updated").Transform(convertJiraTime),
			},

			// JSON fields
			{
				Name:        "components",
				Description: "List of components associated with the issue.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Fields.Components").Transform(extractComponentIds),
			},
			{
				Name:        "fields",
				Description: "Json object containing important subfields of the issue.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "tags",
				Type:        proto.ColumnType_JSON,
				Description: "A map of label names associated with this issue, in Steampipe standard format.",
				Transform:   transform.From(getIssueTags),
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

func listIssues(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	quals := d.QueryContext.GetQuals()

	var sprintId int64
	var epicId int64
	var epicKey string

	jql := ""
	if quals["project_key"] != nil {
		for _, q := range quals["project_key"].Quals {
			op := q.GetStringValue()
			if op != "=" {
				continue
			}
			filterPattern := q.Value.GetStringValue()
			if filterPattern == "" {
				continue
			}
			jql = jql + "project = " + filterPattern
		}
	}

	if quals["project_id"] != nil {
		for _, q := range quals["project_id"].Quals {
			op := q.GetStringValue()
			if op != "=" {
				continue
			}
			filterPattern := q.Value.GetStringValue()
			if filterPattern == "" {
				continue
			}

			if jql == "" {
				jql = jql + "project = " + filterPattern
			} else {
				jql = jql + "&project = " + filterPattern
			}

		}
	}

	if quals["sprint_id"] != nil {
		for _, q := range quals["sprint_id"].Quals {
			op := q.GetStringValue()
			if op != "=" {
				continue
			}
			sprintId = q.Value.GetInt64Value()
			if sprintId == 0 {
				continue
			}

			if jql == "" {
				jql = jql + "sprint = " + strconv.Itoa(int(sprintId))
			} else {
				jql = jql + "&sprint = " + strconv.Itoa(int(sprintId))
			}
		}
	}

	if quals["epic_key"] != nil {
		for _, q := range quals["epic_key"].Quals {
			op := q.GetStringValue()
			if op != "=" {
				continue
			}
			epicKey := q.Value.GetStringValue()
			if epicKey == "" {
				continue
			}

			if jql == "" {
				jql = jql + "\"Epic Link\" = " + epicKey
			} else {
				jql = jql + "&\"Epic Link\" = " + epicKey
			}
		}
	}

	if quals["epic_id"] != nil {
		for _, q := range quals["epic_id"].Quals {
			op := q.GetStringValue()
			if op != "=" {
				continue
			}
			epicId = q.Value.GetInt64Value()
			if epicId == 0 {
				continue
			}

			if jql == "" {
				jql = jql + "\"Epic Link\" = " + strconv.Itoa(int(epicId))
			} else {
				jql = jql + "&\"Epic Link\" = " + strconv.Itoa(int(epicId))
			}
		}
	}

	// plugin.Logger(ctx).Debug("listIssues", "JQL", jql)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	var issues []jira.Issue
	for {
		opt := &jira.SearchOptions{
			MaxResults: 1000, // Max results can go up to 1000
			StartAt:    last,
		}

		// chunk, resp, err := client.Issue.Search("project = "+projectName, opt)
		// chunk, resp, err := client.Issue.Search("\"Epic Link\" = SSP-24", opt)
		// chunk, resp, err := client.Issue.Search("sprint = 1", opt)
		chunk, resp, err := client.Issue.Search(jql, opt)
		if err != nil {
			if isNotFoundError(err) || strings.Contains(err.Error(), "400") {
				return nil, nil
			}
			return nil, err
		}

		total := resp.Total
		if issues == nil {
			issues = make([]jira.Issue, 0, total)
		}

		for _, issue := range chunk {
			d.StreamListItem(ctx, IssueInfo{issue, epicId, epicKey, sprintId})
		}
		last = resp.StartAt + len(chunk)
		if last >= total {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTION

func getIssue(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	issueId := d.KeyColumnQuals["id"].GetStringValue()

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	issue, _, err := client.Issue.Get(issueId, &jira.GetQueryOptions{})
	if err != nil && isNotFoundError(err) {
		return nil, nil
	}

	return IssueInfo{*issue, 0, "", 0}, err
}

//// TRANSFORM FUNCTION

func extractComponentIds(_ context.Context, d *transform.TransformData) (interface{}, error) {
	var componentIds []string
	for _, item := range d.Value.([]*jira.Component) {
		componentIds = append(componentIds, item.ID)
	}
	return componentIds, nil
}

func getIssueTags(_ context.Context, d *transform.TransformData) (interface{}, error) {
	issue := d.HydrateItem.(IssueInfo)

	tags := make(map[string]bool)
	if issue.Fields != nil && issue.Fields.Labels != nil {
		for _, i := range issue.Fields.Labels {
			tags[i] = true
		}
	}
	return tags, nil
}

//// custom struct

type IssueInfo struct {
	jira.Issue
	EpicID   int64
	EpicKey  string
	SprintID int64
}
