package actions

import (
	"github.com/nyaruka/goflow/excellent"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/utils"
)

// TypeSaveFlowResult is our type for the save result action
const TypeSaveFlowResult string = "save_flow_result"

// SaveFlowResultAction can be used to save a result for a flow. The result will be available in the context
// for the run as @run.results.[name]. The optional category can be used as a way of categorizing results,
// this can be useful for reporting or analytics.
//
// Both the value and category fields may be templates. A `run_result_changed` event will be created with the
// final values.
//
// ```
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "save_flow_result",
//     "name": "Gender",
//     "value": "m",
//     "category": "Male"
//   }
// ```
//
// @action save_flow_result
type SaveFlowResultAction struct {
	BaseAction
	Name     string `json:"name" validate:"required"`
	Value    string `json:"value" validate:"required"`
	Category string `json:"category"`
}

// Type returns the type of this action
func (a *SaveFlowResultAction) Type() string { return TypeSaveFlowResult }

// Validate validates our action is valid and has all the assets it needs
func (a *SaveFlowResultAction) Validate(assets flows.SessionAssets) error {
	return nil
}

// Execute runs this action
func (a *SaveFlowResultAction) Execute(run flows.FlowRun, step flows.Step, log flows.EventLog) error {
	// get our localized value if any
	template := run.GetText(utils.UUID(a.UUID()), "value", a.Value)
	value, err := excellent.EvaluateTemplateAsString(run.Environment(), run.Context(), template, false)

	// log any error received
	if err != nil {
		log.Add(events.NewErrorEvent(err))
	}

	categoryLocalized := run.GetText(utils.UUID(a.UUID()), "category", a.Category)
	if a.Category == categoryLocalized {
		categoryLocalized = ""
	}

	log.Add(events.NewRunResultChangedEvent(a.Name, value, a.Category, categoryLocalized, step.NodeUUID(), ""))
	return nil
}
