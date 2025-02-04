package tools_test

import (
	"testing"

	"github.com/nyaruka/goflow/excellent"
	"github.com/nyaruka/goflow/excellent/tools"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/utils"

	"github.com/stretchr/testify/assert"
)

type testXObject struct {
	bar types.XNumber
}

func newTestXObject(bar int) *testXObject {
	return &testXObject{bar: types.NewXNumberFromInt(bar)}
}

// Describe returns a representation of this type for error messages
func (v *testXObject) Describe() string { return "test" }

func (v *testXObject) Reduce(env utils.Environment) types.XPrimitive { return v.bar }

func (v *testXObject) Resolve(env utils.Environment, key string) types.XValue {
	return v.bar
}

// ToXJSON is called when this type is passed to @(json(...))
func (v *testXObject) ToXJSON(env utils.Environment) types.XText {
	return types.ResolveKeys(env, v, "bar").ToXJSON(env)
}

var _ types.XValue = &testXObject{}

func TestRefactorTemplate(t *testing.T) {
	testCases := []struct {
		template   string
		refactored string
		hasError   bool
	}{
		{``, ``, false},
		{`Hi @foo`, `Hi @foo`, false},
		{`@(foo)`, `@(foo)`, false},
		{`@( "Hello"+12345.123 )`, `@("Hello" + 12345.123)`, false},
		{`@foo.bar`, `@foo.bar`, false},
		{`@(foo . bar)`, `@(foo.bar)`, false},
		{`@(OR(TRUE, False, Null))`, `@(or(true, false, null))`, false},
		{`@(foo[ 1 ] + foo[ "x" ])`, `@(foo[1] + foo["x"])`, false},
		{`@(-1+( 2/3 )*4^5)`, `@(-1 + (2 / 3) * 4 ^ 5)`, false},
		{`@("x"&"y")`, `@("x" & "y")`, false},
		{`@(AND("x"="y", "x"!="y"))`, `@(and("x" = "y", "x" != "y"))`, false},
		{`@(AND(1>2, 3<4, 5>=6, 7<=8))`, `@(and(1 > 2, 3 < 4, 5 >= 6, 7 <= 8))`, false},
		{`@(FOO_Func(x, y))`, `@(foo_func(x, y))`, false},
		{`@(1 / ) @(1+2)`, `@(1 / ) @(1 + 2)`, true},
	}

	env := utils.NewEnvironmentBuilder().Build()
	vars := types.NewXMap(map[string]types.XValue{
		"foo": newTestXObject(123),
	})
	topLevels := []string{"foo"}

	for _, tc := range testCases {
		actual, err := tools.RefactorTemplate(tc.template, topLevels)

		assert.Equal(t, tc.refactored, actual, "refactor mismatch for template: %s", tc.template)

		if tc.hasError {
			assert.Error(t, err, "expected error for template: %s", tc.template)
		} else {
			assert.NoError(t, err, "unexpected error for template: %s, err: %s", tc.template, err)

			// test that the original and the refactored template evaluate equally
			originalValue, _ := excellent.EvaluateTemplate(env, vars, tc.template, topLevels)
			refactoredValue, _ := excellent.EvaluateTemplate(env, vars, actual, topLevels)

			assert.Equal(t, originalValue, refactoredValue, "refactoring of template %s gives different value: %s", tc.template, refactoredValue)
		}
	}
}
