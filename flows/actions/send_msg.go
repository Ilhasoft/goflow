package actions

import (
	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
)

// TypeSendMsg is the type for msg actions
const TypeSendMsg string = "send_msg"

// SendMsgAction can be used to send a message to one or more contacts. It accepts a list of URNs, a list of groups
// and a list of contacts.
//
// The URNs and text fields may be templates. A `send_msg` event will be created for each unique urn, contact and group
// with the evaluated text.
//
// ```
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "send_msg",
//     "urns": ["tel:+12065551212"],
//     "text": "Hi @contact.name, are you ready to complete today's survey?"
//   }
// ```
//
// @action send_msg
type SendMsgAction struct {
	BaseAction
	Text         string                    `json:"text"`
	Attachments  []string                  `json:"attachments"`
	QuickReplies []string                  `json:"quick_replies,omitempty"`
	URNs         []urns.URN                `json:"urns,omitempty"`
	Contacts     []*flows.ContactReference `json:"contacts,omitempty" validate:"dive"`
	Groups       []*flows.GroupReference   `json:"groups,omitempty" validate:"dive"`
	LegacyVars   []string                  `json:"legacy_vars,omitempty"`
}

// Type returns the type of this action
func (a *SendMsgAction) Type() string { return TypeSendMsg }

// Validate validates our action is valid and has all the assets it needs
func (a *SendMsgAction) Validate(assets flows.SessionAssets) error {
	return a.validateGroups(assets, a.Groups)
}

// Execute runs this action
func (a *SendMsgAction) Execute(run flows.FlowRun, step flows.Step, log flows.EventLog) error {
	evaluatedText, evaluatedAttachments, evaluatedQuickReplies := a.evaluateMessage(run, step, a.Text, a.Attachments, a.QuickReplies, log)

	urnList, contactRefs, groupRefs, err := a.resolveContactsAndGroups(run, step, a.URNs, a.Contacts, a.Groups, a.LegacyVars, log)
	if err != nil {
		return err
	}

	log.Add(events.NewSendMsgEvent(evaluatedText, evaluatedAttachments, evaluatedQuickReplies, urnList, contactRefs, groupRefs))

	return nil
}
