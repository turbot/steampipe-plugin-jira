package jira

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableWorkflow(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "jira_workflow",
		Description: "A Jira workflow is a set of statuses and transitions that an issue moves through during its lifecycle, and typically represents a process within your organization.",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("name"),
			Hydrate:    getWorkflow,
		},
		List: &plugin.ListConfig{
			Hydrate: listWorkflows,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the workflow.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID.Name"),
			},
			{
				Name:        "entity_id",
				Description: "The entity ID of the workflow.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID.EntityID"),
			},
			{
				Name:        "description",
				Description: "The description of the workflow.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_default",
				Description: "Whether this is the default workflow.",
				Type:        proto.ColumnType_BOOL,
			},

			// json fields
			{
				Name:        "transitions",
				Description: "The transitions of the workflow.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "statuses",
				Description: "The statuses of the workflow.",
				Type:        proto.ColumnType_JSON,
			},

			// Standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID.Name"),
			},
		},
	}
}

//// LIST FUNCTION

func listWorkflows(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("listWorkflows")

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	last := 0
	maxResults := 1000
	for {
		apiEndpoint := fmt.Sprintf(
			"/rest/api/3/workflow/search?startAt=%d&maxResults=%d&expand=transitions,transitions.rules,statuses,statuses.properties,default",
			last,
			maxResults,
		)

		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return nil, err
		}

		listResult := new(ListWorkflowResult)
		_, err = client.Do(req, listResult)
		if err != nil {
			logger.Error("listWorkflows", "Error", err)
			return nil, err
		}

		for _, workflow := range listResult.Values {
			d.StreamListItem(ctx, workflow)
		}

		last = listResult.StartAt + len(listResult.Values)
		if listResult.IsLast {
			return nil, nil
		}
	}
}

//// HYDRATE FUNCTIONS

func getWorkflow(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getWorkflow")

	workflowName := strings.Replace(d.KeyColumnQuals["name"].GetStringValue(), " ", "%20", -1)

	client, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf(
		"/rest/api/3/workflow/search?workflowName=%s&expand=transitions,transitions.rules,statuses,statuses.properties,default",
		workflowName,
	)
	logger.Trace("getWorkflow", "API Endpoint", apiEndpoint)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	workflow := new(ListWorkflowResult)
	_, err = client.Do(req, workflow)
	if err != nil {
		logger.Error("listWorkflows", "Error", err)
		return nil, err
	}
	if len(workflow.Values) < 1 {
		return nil, nil
	}
	logger.Trace("getWorkflow", "Output", workflow)

	return workflow.Values[0], nil
}

//// Custom Structs

type ListWorkflowResult struct {
	Self       string     `json:"self"`
	NextPage   string     `json:"nextPage"`
	MaxResults int        `json:"maxResults"`
	StartAt    int        `json:"startAt"`
	Total      int        `json:"total"`
	IsLast     bool       `json:"isLast"`
	Values     []Workflow `json:"values"`
}

type Workflow struct {
	ID          WorkflowID           `json:"id"`
	Description string               `json:"description"`
	Transitions []WorkflowTransition `json:"transitions"` // Check fields
	Statuses    []WorkflowStatus     `json:"statuses"`
	IsDefault   bool                 `json:"isDefault"`
}

type WorkflowID struct {
	Name     string `json:"name"`
	EntityID string `json:"entityId"`
}

type WorkflowStatus struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Properties WorkflowStatusProperty `json:"properties"`
}

type WorkflowStatusProperty struct {
	IssueEditable bool `json:"issueEditable"`
}

type WorkflowTransition struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	From        []string                 `json:"from"`
	To          string                   `json:"to"`
	Type        string                   `json:"type"`
	Screen      WorkflowTransitionScreen `json:"screen"`
	Rules       WorkflowRules            `json:"rules"`
}

type WorkflowTransitionScreen struct {
	ID string `json:"id"`
}

type WorkflowRules struct {
	ConditionsTree interface{}               `json:"conditionsTree"`
	Validators     []WorkflowTransitionRules `json:"validators"`
	PostFunctions  []WorkflowTransitionRules `json:"postFunctions"`
}

type WorkflowTransitionRules struct {
	Type          string      `json:"type"`
	Configuration interface{} `json:"configuration"`
}
