package actions

import (
	"fmt"
	"strings"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/utils"
)

// TypeSetContactLanguage is the type for the set contact Language action
const TypeSetContactLanguage string = "set_contact_language"

// SetContactLanguageAction can be used to update the name of the contact. A `contact_language_changed`
// event will be created with the corresponding value.
//
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "set_contact_language",
//     "language": "eng"
//   }
//
// @action set_contact_language
type SetContactLanguageAction struct {
	BaseAction
	Language string `json:"language"`
}

// Type returns the type of this action
func (a *SetContactLanguageAction) Type() string { return TypeSetContactLanguage }

// Validate validates our action is valid and has all the assets it needs
func (a *SetContactLanguageAction) Validate(assets flows.SessionAssets) error {
	// check language is valid if specified
	if a.Language != "" {
		if _, err := utils.ParseLanguage(a.Language); err != nil {
			return err
		}
	}
	return nil
}

// Execute runs this action
func (a *SetContactLanguageAction) Execute(run flows.FlowRun, step flows.Step, log flows.EventLog) error {
	if run.Contact() == nil {
		log.Add(events.NewFatalErrorEvent(fmt.Errorf("can't execute action in session without a contact")))
		return nil
	}

	// get our localized value if any
	template := run.GetText(utils.UUID(a.UUID()), "language", a.Language)
	language, err := run.EvaluateTemplateAsString(template, false)
	language = strings.TrimSpace(language)

	// if we received an error, log it
	if err != nil {
		log.Add(events.NewErrorEvent(err))
		return nil
	}

	log.Add(events.NewContactLanguageChangedEvent(language))
	return nil
}
