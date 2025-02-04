package routers

import (
	"strings"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/routers/tests"
	"github.com/nyaruka/goflow/utils"

	"github.com/pkg/errors"
)

func init() {
	RegisterType(TypeSwitch, func() flows.Router { return &SwitchRouter{} })
}

// TypeSwitch is the constant for our switch router
const TypeSwitch string = "switch"

// Case represents a single case and test in our switch
type Case struct {
	UUID        utils.UUID     `json:"uuid"                 validate:"required"`
	Type        string         `json:"type"                 validate:"required"`
	Arguments   []string       `json:"arguments,omitempty"`
	OmitOperand bool           `json:"omit_operand,omitempty"`
	ExitUUID    flows.ExitUUID `json:"exit_uuid"            validate:"required"`
}

// NewCase creates a new case
func NewCase(uuid utils.UUID, type_ string, arguments []string, omitOperand bool, exitUUID flows.ExitUUID) *Case {
	return &Case{
		UUID:        uuid,
		Type:        type_,
		Arguments:   arguments,
		OmitOperand: omitOperand,
		ExitUUID:    exitUUID,
	}
}

// LocalizationUUID gets the UUID which identifies this object for localization
func (c *Case) LocalizationUUID() utils.UUID { return utils.UUID(c.UUID) }

// Inspect inspects this object and any children
func (c *Case) Inspect(inspect func(flows.Inspectable)) {
	inspect(c)
}

// EnumerateTemplates enumerates all expressions on this object and its children
func (c *Case) EnumerateTemplates(localization flows.Localization, include func(string)) {
	for _, arg := range c.Arguments {
		include(arg)
	}

	flows.EnumerateTemplateTranslations(localization, c, "arguments", include)
}

// RewriteTemplates rewrites all templates on this object and its children
func (c *Case) RewriteTemplates(localization flows.Localization, rewrite func(string) string) {
	for a := range c.Arguments {
		c.Arguments[a] = rewrite(c.Arguments[a])
	}

	flows.RewriteTemplateTranslations(localization, c, "arguments", rewrite)
}

// EnumerateDependencies enumerates all dependencies on this object and its children
func (c *Case) EnumerateDependencies(localization flows.Localization, include func(assets.Reference)) {
	// currently only the HAS_GROUP router test can produce a dependency
	if c.Type == "has_group" && len(c.Arguments) > 0 {
		include(assets.NewGroupReference(assets.GroupUUID(c.Arguments[0]), ""))

		// the group UUID might be different in different translations
		for _, lang := range localization.Languages() {
			arguments := localization.GetTranslations(lang).GetTextArray(c.UUID, "arguments")
			if len(arguments) > 0 {
				include(assets.NewGroupReference(assets.GroupUUID(arguments[0]), ""))
			}
		}
	}
}

// EnumerateResultNames enumerates all result names on this object
func (c *Case) EnumerateResultNames(include func(string)) {}

// SwitchRouter is a router which allows specifying 0-n cases which should each be tested in order, following
// whichever case returns true, or if none do, then taking the default exit
type SwitchRouter struct {
	BaseRouter
	Default flows.ExitUUID `json:"default_exit_uuid"   validate:"omitempty,uuid4"`
	Operand string         `json:"operand"             validate:"required"`
	Cases   []*Case        `json:"cases"`
}

// NewSwitchRouter creates a new switch router
func NewSwitchRouter(defaultExit flows.ExitUUID, operand string, cases []*Case, resultName string) *SwitchRouter {
	return &SwitchRouter{
		BaseRouter: newBaseRouter(TypeSwitch, resultName),
		Default:    defaultExit,
		Operand:    operand,
		Cases:      cases,
	}
}

// Validate validates the arguments for this router
func (r *SwitchRouter) Validate(exits []flows.Exit) error {
	// helper to look for the given exit UUID
	hasExit := func(exitUUID flows.ExitUUID) bool {
		found := false
		for _, e := range exits {
			if e.UUID() == exitUUID {
				found = true
				break
			}
		}
		return found
	}

	if r.Default != "" && !hasExit(r.Default) {
		return errors.Errorf("default exit %s is not a valid exit", r.Default)
	}

	for _, c := range r.Cases {
		if !hasExit(c.ExitUUID) {
			return errors.Errorf("case exit %s is not a valid exit", c.ExitUUID)
		}
	}

	return nil
}

// PickRoute evaluates each of the tests on our cases in order, returning the exit for the first case which
// evaluates to a true. If no cases evaluate to true, then the default exit (if specified) is returned
func (r *SwitchRouter) PickRoute(run flows.FlowRun, exits []flows.Exit, step flows.Step) (*string, flows.Route, error) {
	env := run.Environment()

	// first evaluate our operand
	operand, err := run.EvaluateTemplateValue(r.Operand)
	if err != nil {
		run.LogError(step, err)
	}

	var operandAsStr *string
	if operand != nil {
		asText, _ := types.ToXText(env, operand)
		asString := asText.Native()
		operandAsStr = &asString
	}

	// each of our cases
	for _, c := range r.Cases {
		test := strings.ToLower(c.Type)

		// try to look up our function
		xtest := tests.XTESTS[test]
		if xtest == nil {
			return nil, flows.NoRoute, errors.Errorf("unknown test '%s', taking no exit", c.Type)
		}

		// build our argument list
		args := make([]types.XValue, 0, 1)
		if !c.OmitOperand {
			args = append(args, operand)
		}

		localizedArgs := run.GetTextArray(c.UUID, "arguments", c.Arguments)
		for i := range c.Arguments {
			test := localizedArgs[i]
			arg, err := run.EvaluateTemplateValue(test)
			if err != nil {
				run.LogError(step, err)
			}
			args = append(args, arg)
		}

		// call our function
		result := xtest(env, args...)

		// tests have to return either errors or test results
		switch typedResult := result.(type) {
		case types.XError:
			// test functions can return an error
			run.LogError(step, errors.Errorf("error calling test %s: %s", strings.ToUpper(test), typedResult.Error()))
			continue
		case tests.XTestResult:
			// looks truthy, lets return this exit
			if typedResult.Matched() {
				resultAsStr, xerr := types.ToXText(env, typedResult.Match())
				if xerr != nil {
					return nil, flows.NoRoute, xerr
				}

				return operandAsStr, flows.NewRoute(c.ExitUUID, resultAsStr.Native(), typedResult.Extra()), nil
			}
		default:
			return nil, flows.NoRoute, errors.Errorf("unexpected result type from test %v: %#v", xtest, result)
		}
	}

	// we have a default exit, use that
	if r.Default != "" {
		// evaluate our operand as a string
		value, xerr := types.ToXText(env, operand)
		if xerr != nil {
			run.LogError(step, xerr)
		}

		return operandAsStr, flows.NewRoute(r.Default, value.Native(), nil), nil
	}

	// no matches, no defaults, no route
	return operandAsStr, flows.NoRoute, nil
}

// Inspect inspects this object and any children
func (r *SwitchRouter) Inspect(inspect func(flows.Inspectable)) {
	inspect(r)

	for _, cs := range r.Cases {
		cs.Inspect(inspect)
	}
}

// EnumerateTemplates enumerates all expressions on this object and its children
func (r *SwitchRouter) EnumerateTemplates(localization flows.Localization, include func(string)) {
	include(r.Operand)
}

// RewriteTemplates rewrites all templates on this object and its children
func (r *SwitchRouter) RewriteTemplates(localization flows.Localization, rewrite func(string) string) {
	r.Operand = rewrite(r.Operand)
}

// EnumerateDependencies enumerates all dependencies on this object and its children
func (r *SwitchRouter) EnumerateDependencies(localization flows.Localization, include func(assets.Reference)) {
}
