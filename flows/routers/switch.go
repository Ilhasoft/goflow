package routers

import (
	"fmt"
	"strings"

	"github.com/nyaruka/goflow/excellent"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/utils"
)

const SWITCH string = "switch"

type Case struct {
	UUID      flows.UUID     `json:"uuid"        validate:"nonzero"`
	Type      string         `json:"type"        validate:"nonzero"`
	Arguments []string       `json:"arguments"`
	Exit      flows.ExitUUID `json:"exit"        validate:"nonzero"`
}

type SwitchRouter struct {
	Default flows.ExitUUID `json:"default"`
	Operand string         `json:"operand"     validate:"nonzero"`
	Cases   []Case         `json:"cases"`
	BaseRouter
}

func (r *SwitchRouter) Type() string { return SWITCH }

func (r *SwitchRouter) Validate(exits []flows.Exit) error {
	err := utils.ValidateAll(r)
	for _, c := range r.Cases {
		err = utils.ValidateAll(c)
		if err != nil {
			return err
		}

		// find the matching exit
		found := false
		for _, e := range exits {
			if e.UUID() == c.Exit {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Exit '%s' missing from node", c.Exit)
		}
	}

	return err
}

func (r *SwitchRouter) PickRoute(run flows.FlowRun, exits []flows.Exit, step flows.Step) (flows.Route, error) {
	env := run.Environment()

	// first evaluate our operand
	operand, err := excellent.EvaluateTemplate(env, run.Context(), r.Operand)
	if err != nil {
		run.AddError(step, err)
	}

	// each of our cases
	for _, c := range r.Cases {
		test := strings.ToLower(c.Type)

		// try to look up our function
		xtest := excellent.XTESTS[test]
		if xtest == nil {
			run.AddError(step, fmt.Errorf("Unknown test '%s', taking no exit", c.Type))
			return flows.NoRoute, nil
		}

		// build our argument list
		args := make([]interface{}, len(c.Arguments)+1)
		args[0] = operand
		for i := range c.Arguments {
			test := run.GetText(c.UUID, fmt.Sprintf("args.%d", i), c.Arguments[i])
			args[i+1], err = excellent.EvaluateTemplate(env, run.Context(), test)
			if err != nil {
				run.AddError(step, err)
			}
		}

		// call our function
		rawResult := xtest(env, args...)
		err, isErr := rawResult.(error)
		if isErr {
			return flows.NoRoute, err
		}

		// ok, not an error, must be an XTestResult
		result, isResult := rawResult.(excellent.XTestResult)
		if !isResult {
			return flows.NoRoute, fmt.Errorf("Unexpected result type from test %v: %#v", xtest, result)
		}

		// looks truthy, lets return this exit
		if result.Matched() {
			asStr, err := utils.ToString(env, result.Match())
			if err != nil {
				return flows.NoRoute, err
			}

			return flows.NewRoute(c.Exit, asStr), nil
		}
	}

	// we have a default exit, use that
	if r.Default != "" {
		return flows.NewRoute(r.Default, "default"), nil
	}

	// no matches, no defaults, no route
	return flows.NoRoute, nil
}
