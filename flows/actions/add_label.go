package actions

import (
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
)

// TypeAddLabel is our type for add label actions
const TypeAddLabel string = "add_label"

// AddLabelAction can be used to add a label to the last user input on a flow. An `add_label` event
// will be created with the input UUID and label UUIDs when this action is encountered. If there is
// no user input at that point then this action will be ignored.
//
// ```
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "add_label",
//     "labels": [{
//       "uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
//       "name": "complaint"
//     }]
//   }
// ```
//
// @action add_label
type AddLabelAction struct {
	BaseAction
	Labels []*flows.LabelReference `json:"labels" validate:"required,min=1,dive"`
}

// Type returns the type of this action
func (a *AddLabelAction) Type() string { return TypeAddLabel }

// Validate validates our action is valid and has all the assets it needs
func (a *AddLabelAction) Validate(assets flows.SessionAssets) error {
	// check we have all labels
	return a.validateLabels(assets, a.Labels)
}

// Execute runs the labeling action
func (a *AddLabelAction) Execute(run flows.FlowRun, step flows.Step, log flows.EventLog) error {
	// only generate event if run has input
	input := run.Input()
	if input == nil {
		return nil
	}

	labels, err := a.resolveLabels(run, step, a.Labels, log)
	if err != nil {
		return err
	}

	labelRefs := make([]*flows.LabelReference, 0, len(labels))
	for _, label := range labels {
		labelRefs = append(labelRefs, label.Reference())
	}

	if len(labelRefs) > 0 {
		log.Add(events.NewLabelAddedEvent(input.UUID(), labelRefs))
	}

	return nil
}
