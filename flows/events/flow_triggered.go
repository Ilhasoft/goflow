package events

import (
	"github.com/nyaruka/goflow/flows"
)

// TypeFlowTriggered is the type of our flow triggered event
const TypeFlowTriggered string = "flow_triggered"

// FlowTriggeredEvent events are created when an action wants to start a subflow
//
// ```
//   {
//     "type": "flow_triggered",
//     "created_on": "2006-01-02T15:04:05Z",
//     "flow": {"uuid": "0e06f977-cbb7-475f-9d0b-a0c4aaec7f6a", "name": "Registration"},
//     "parent_run_uuid": "95eb96df-461b-4668-b168-727f8ceb13dd"
//   }
// ```
//
// @event flow_triggered
type FlowTriggeredEvent struct {
	BaseEvent
	Flow          *flows.FlowReference `json:"flow" validate:"required"`
	ParentRunUUID flows.RunUUID        `json:"parent_run_uuid" validate:"omitempty,uuid4"`
}

// NewFlowTriggeredEvent returns a new flow triggered event for the passed in flow and parent run
func NewFlowTriggeredEvent(flow *flows.FlowReference, parentRunUUID flows.RunUUID) *FlowTriggeredEvent {
	return &FlowTriggeredEvent{
		BaseEvent:     NewBaseEvent(),
		Flow:          flow,
		ParentRunUUID: parentRunUUID,
	}
}

// Type returns the type of this event
func (e *FlowTriggeredEvent) Type() string { return TypeFlowTriggered }

// AllowedOrigin determines where this event type can originate
func (e *FlowTriggeredEvent) AllowedOrigin() flows.EventOrigin { return flows.EventOriginEngine }

// Validate validates our event is valid and has all the assets it needs
func (e *FlowTriggeredEvent) Validate(assets flows.SessionAssets) error {
	return nil
}

// Apply applies this event to the given run
func (e *FlowTriggeredEvent) Apply(run flows.FlowRun) error {
	flow, err := run.Session().Assets().GetFlow(e.Flow.UUID)
	if err != nil {
		return err
	}

	parentRun, err := run.Session().GetRun(e.ParentRunUUID)
	if err != nil {
		return err
	}

	run.Session().PushFlow(flow, parentRun)
	return nil
}
