package events

import "github.com/nyaruka/goflow/flows"

// TypeEmailSent is our type for the email event
const TypeEmailSent string = "email_sent"

// EmailSentEvent events are created for each recipient which should receive an email.
//
// ```
//   {
//     "type": "email_sent",
//     "created_on": "2006-01-02T15:04:05Z",
//     "addresses": ["foo@bar.com"],
//     "subject": "Your activation token",
//     "body": "Your activation token is AAFFKKEE"
//   }
// ```
//
// @event email_sent
type EmailSentEvent struct {
	BaseEvent
	Addresses []string `json:"addresses" validate:"required,min=1"`
	Subject   string   `json:"subject" validate:"required"`
	Body      string   `json:"body"`
}

// NewEmailSentEvent returns a new email event with the passed in subject, body and emails
func NewEmailSentEvent(addresses []string, subject string, body string) *EmailSentEvent {
	return &EmailSentEvent{
		BaseEvent: NewBaseEvent(),
		Addresses: addresses,
		Subject:   subject,
		Body:      body,
	}
}

// Type returns the type of this event
func (a *EmailSentEvent) Type() string { return TypeEmailSent }

// Apply applies this event to the given run
func (e *EmailSentEvent) Apply(run flows.FlowRun) error {
	return nil
}
