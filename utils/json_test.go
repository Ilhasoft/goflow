package utils_test

import (
	"testing"

	"github.com/nyaruka/goflow/utils"
)

func TestJSON(t *testing.T) {
	var jsonTests = []struct {
		JSON     []byte
		lookup   string
		expected string
	}{
		{nil, "key", ""},
		{[]byte(`malformed`), "key", ""},
		{[]byte(`["one", "two", "three"]`), "0", "one"},
		{[]byte(`["escaped \"string\""]`), "0", `escaped "string"`},
		{[]byte(`{"1": "one"}`), "1", "one"},
		{[]byte(`{"arr": ["one", "two"]}`), "arr[1]", "two"},
		{[]byte(`{"arr": ["one", "two"]}`), "arr.1", "two"},
		{[]byte(`{"key": {"key2": "val2"}}`), "key.key2", "val2"},
		{[]byte(`{"key": {"key-with-dash": "val2"}}`), `key["key-with-dash"]`, "val2"},
		{[]byte(`{"key": {"key with space": "val2"}}`), `key["key with space"]`, "val2"},
	}

	env := utils.NewDefaultEnvironment()
	for _, test := range jsonTests {
		fragment := utils.JSONFragment(test.JSON)
		value := utils.ResolveVariable(env, fragment, test.lookup)

		valueStr, _ := utils.ToString(env, value)
		if valueStr != test.expected {
			t.Errorf("Expected: '%s' Got: '%s' for lookup: '%s' and JSON fragment:\n%s", test.expected, valueStr, test.lookup, test.JSON)
		}
	}
}
