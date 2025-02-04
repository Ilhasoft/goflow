package actions

import (
	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/utils"
)

func init() {
	RegisterType(TypeSendBroadcast, func() flows.Action { return &SendBroadcastAction{} })
}

// TypeSendBroadcast is the type for the send broadcast action
const TypeSendBroadcast string = "send_broadcast"

// SendBroadcastAction can be used to send a message to one or more contacts. It accepts a list of URNs, a list of groups
// and a list of contacts.
//
// The URNs and text fields may be templates. A [event:broadcast_created] event will be created for each unique urn, contact and group
// with the evaluated text.
//
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "send_broadcast",
//     "urns": ["tel:+12065551212"],
//     "text": "Hi @contact.name, are you ready to complete today's survey?"
//   }
//
// @action send_broadcast
type SendBroadcastAction struct {
	BaseAction
	onlineAction
	otherContactsAction
	createMsgAction
}

// NewSendBroadcastAction creates a new send broadcast action
func NewSendBroadcastAction(uuid flows.ActionUUID, text string, attachments []string, quickReplies []string, urns []urns.URN, contacts []*flows.ContactReference, groups []*assets.GroupReference, legacyVars []string) *SendBroadcastAction {
	return &SendBroadcastAction{
		BaseAction: NewBaseAction(TypeSendBroadcast, uuid),
		otherContactsAction: otherContactsAction{
			URNs:       urns,
			Contacts:   contacts,
			Groups:     groups,
			LegacyVars: legacyVars,
		},
		createMsgAction: createMsgAction{
			Text:         text,
			Attachments:  attachments,
			QuickReplies: quickReplies,
		},
	}
}

// Execute runs this action
func (a *SendBroadcastAction) Execute(run flows.FlowRun, step flows.Step, logModifier flows.ModifierCallback, logEvent flows.EventCallback) error {
	urnList, contactRefs, groupRefs, err := a.resolveContactsAndGroups(run, a.URNs, a.Contacts, a.Groups, a.LegacyVars, logEvent)
	if err != nil {
		return err
	}

	translations := make(map[utils.Language]*events.BroadcastTranslation)
	languages := append([]utils.Language{run.Flow().Language()}, run.Flow().Localization().Languages()...)

	// evaluate the broadcast in each language we have translations for
	for _, language := range languages {
		languages := []utils.Language{language, run.Flow().Language()}

		evaluatedText, evaluatedAttachments, evaluatedQuickReplies := a.evaluateMessage(run, languages, a.Text, a.Attachments, a.QuickReplies, logEvent)
		translations[language] = &events.BroadcastTranslation{
			Text:         evaluatedText,
			Attachments:  evaluatedAttachments,
			QuickReplies: evaluatedQuickReplies,
		}
	}

	logEvent(events.NewBroadcastCreatedEvent(translations, run.Flow().Language(), urnList, contactRefs, groupRefs))

	return nil
}

// Inspect inspects this object and any children
func (a *SendBroadcastAction) Inspect(inspect func(flows.Inspectable)) {
	inspect(a)

	for _, g := range a.Groups {
		flows.InspectReference(g, inspect)
	}
	for _, c := range a.Contacts {
		flows.InspectReference(c, inspect)
	}
}

// EnumerateTemplates enumerates all expressions on this object and its children
func (a *SendBroadcastAction) EnumerateTemplates(localization flows.Localization, include func(string)) {
	include(a.Text)
	flows.EnumerateTemplateArray(a.Attachments, include)
	flows.EnumerateTemplateArray(a.QuickReplies, include)
	flows.EnumerateTemplateTranslations(localization, a, "text", include)
	flows.EnumerateTemplateTranslations(localization, a, "attachments", include)
	flows.EnumerateTemplateTranslations(localization, a, "quick_replies", include)
	flows.EnumerateTemplateArray(a.LegacyVars, include)
}

// RewriteTemplates rewrites all templates on this object and its children
func (a *SendBroadcastAction) RewriteTemplates(localization flows.Localization, rewrite func(string) string) {
	a.Text = rewrite(a.Text)
	flows.RewriteTemplateArray(a.Attachments, rewrite)
	flows.RewriteTemplateArray(a.QuickReplies, rewrite)
	flows.RewriteTemplateTranslations(localization, a, "text", rewrite)
	flows.RewriteTemplateTranslations(localization, a, "attachments", rewrite)
	flows.RewriteTemplateTranslations(localization, a, "quick_replies", rewrite)
	flows.RewriteTemplateArray(a.LegacyVars, rewrite)
}
