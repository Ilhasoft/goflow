package actions_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/actions"
	"github.com/nyaruka/goflow/flows/triggers"
	"github.com/nyaruka/goflow/test"
	"github.com/nyaruka/goflow/utils"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var contactJSON = `{
	"uuid": "5d76d86b-3bb9-4d5a-b822-c9d86f5d8e4f",
	"name": "Ryan Lewis",
	"language": "eng",
	"timezone": "America/Guayaquil",
	"urns": [],
	"groups": [
		{"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d", "name": "Testers"},
		{"uuid": "0ec97956-c451-48a0-a180-1ce766623e31", "name": "Males"}
	],
	"fields": {
		"gender": {
			"text": "Male"
		}
	},
	"created_on": "2018-06-20T11:40:30.123456789-00:00"
}`

func TestActionTypes(t *testing.T) {
	assetsJSON, err := ioutil.ReadFile("testdata/_assets.json")
	require.NoError(t, err)

	server := test.NewTestHTTPServer(49996)

	for typeName := range actions.RegisteredTypes() {
		testActionType(t, assetsJSON, typeName, server.URL)
	}
}

type inspectionResults struct {
	Templates    []string `json:"templates"`
	Dependencies []string `json:"dependencies"`
	ResultNames  []string `json:"result_names"`
}

func testActionType(t *testing.T, assetsJSON json.RawMessage, typeName string, testServerURL string) {
	testFile, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", typeName))
	require.NoError(t, err)

	tests := []struct {
		Description     string             `json:"description"`
		NoContact       bool               `json:"no_contact"`
		NoURNs          bool               `json:"no_urns"`
		NoInput         bool               `json:"no_input"`
		Action          json.RawMessage    `json:"action"`
		ValidationError string             `json:"validation_error"`
		Events          []json.RawMessage  `json:"events"`
		ContactAfter    json.RawMessage    `json:"contact_after"`
		Inspection      *inspectionResults `json:"inspection"`
	}{}

	err = json.Unmarshal(testFile, &tests)
	require.NoError(t, err)

	defer utils.SetTimeSource(utils.DefaultTimeSource)
	defer utils.SetUUIDGenerator(utils.DefaultUUIDGenerator)

	for _, tc := range tests {
		utils.SetTimeSource(utils.NewFixedTimeSource(time.Date(2018, 10, 18, 14, 20, 30, 123456, time.UTC)))
		utils.SetUUIDGenerator(utils.NewSeededUUID4Generator(12345))

		testName := fmt.Sprintf("test '%s' for action type '%s'", tc.Description, typeName)

		// create unstarted session from our assets
		session, err := test.CreateSession(assetsJSON, testServerURL)
		require.NoError(t, err)

		// read the action to be tested
		action, err := actions.ReadAction(tc.Action)
		require.NoError(t, err, "error loading action in %s", testName)
		assert.Equal(t, typeName, action.Type())

		// get a suitable "holder" flow
		var flowUUID assets.FlowUUID
		if len(action.AllowedFlowTypes()) == 1 && action.AllowedFlowTypes()[0] == flows.FlowTypeVoice {
			flowUUID = assets.FlowUUID("7a84463d-d209-4d3e-a0ff-79f977cd7bd0")
		} else {
			flowUUID = assets.FlowUUID("bead76f5-dac4-4c9d-996c-c62b326e8c0a")
		}

		flow, err := session.Assets().Flows().Get(flowUUID)
		require.NoError(t, err)

		// if not, add it to our flow
		flow.Nodes()[0].AddAction(action)

		// if this action is expected to cause flow validation failure, check that
		err = flow.Validate(session.Assets())
		if tc.ValidationError != "" {
			rootErr := errors.Cause(err)
			assert.EqualError(t, rootErr, tc.ValidationError, "validation error mismatch in %s", testName)
			continue
		} else {
			assert.NoError(t, err, "unexpected validation error in %s", testName)
		}

		// optionally load our contact
		var contact *flows.Contact
		if !tc.NoContact {
			contact, err = flows.ReadContact(session.Assets(), json.RawMessage(contactJSON), assets.PanicOnMissing)
			require.NoError(t, err)

			// optionally give our contact some URNs
			if !tc.NoURNs {
				channel := session.Assets().Channels().Get("57f1078f-88aa-46f4-a59a-948a5739c03d")
				contact.AddURN(flows.NewContactURN(urns.URN("tel:+12065551212?channel=57f1078f-88aa-46f4-a59a-948a5739c03d&id=123"), channel))
				contact.AddURN(flows.NewContactURN(urns.URN("twitterid:54784326227#nyaruka"), nil))
			}
		}

		var trigger flows.Trigger
		ignoreEventCount := 0
		if tc.NoInput {
			var connection *flows.Connection
			if flow.Type() == flows.FlowTypeVoice {
				channel := session.Assets().Channels().Get("57f1078f-88aa-46f4-a59a-948a5739c03d")
				connection = flows.NewConnection(channel.Reference(), urns.URN("tel:+12065551212"))
				trigger = triggers.NewManualVoiceTrigger(utils.NewEnvironmentBuilder().Build(), flow.Reference(), contact, connection, nil)
			} else {
				trigger = triggers.NewManualTrigger(utils.NewEnvironmentBuilder().Build(), flow.Reference(), contact, nil)
			}
		} else {
			msg := flows.NewMsgIn(flows.MsgUUID("aa90ce99-3b4d-44ba-b0ca-79e63d9ed842"), urns.URN("tel:+12065551212"), nil, "Hi everybody", nil)
			trigger = triggers.NewMsgTrigger(utils.NewEnvironmentBuilder().Build(), flow.Reference(), contact, msg, nil)
			ignoreEventCount = 1 // need to ignore the msg_received event this trigger creates
		}

		_, err = session.Start(trigger)
		require.NoError(t, err)

		// check events are what we expected
		run := session.Runs()[0]
		runEvents := run.Events()
		actualEventsJSON, _ := json.Marshal(runEvents[ignoreEventCount:])
		expectedEventsJSON, _ := json.Marshal(tc.Events)
		test.AssertEqualJSON(t, expectedEventsJSON, actualEventsJSON, "events mismatch in %s", testName)

		// check contact is in the expected state
		if tc.ContactAfter != nil {
			contactJSON, _ := json.Marshal(session.Contact())

			test.AssertEqualJSON(t, tc.ContactAfter, contactJSON, "contact mismatch in %s", testName)
		}

		// try marshaling the action back to JSON
		actionJSON, err := json.Marshal(action)
		test.AssertEqualJSON(t, tc.Action, actionJSON, "marshal mismatch in %s", testName)

		// finally try inspecting this action
		if tc.Inspection != nil {
			templates := flow.ExtractTemplates()
			assert.Equal(t, tc.Inspection.Templates, templates, "inspected templates mismatch in %s", testName)

			dependencies := flow.ExtractDependencies()
			depStrings := make([]string, len(dependencies))
			for i := range dependencies {
				depStrings[i] = dependencies[i].String()
			}
			assert.Equal(t, tc.Inspection.Dependencies, depStrings, "inspected dependencies mismatch in %s", testName)

			resultNames := flow.ExtractResultNames()
			assert.Equal(t, tc.Inspection.ResultNames, resultNames, "inspected result names mismatch in %s", testName)
		}
	}
}

func TestConstructors(t *testing.T) {
	actionUUID := flows.ActionUUID("ad154980-7bf7-4ab8-8728-545fd6378912")

	tests := []struct {
		action flows.Action
		json   string
	}{
		{
			actions.NewAddContactGroupsAction(
				actionUUID,
				[]*assets.GroupReference{
					assets.NewGroupReference(assets.GroupUUID("b7cf0d83-f1c9-411c-96fd-c511a4cfa86d"), "Testers"),
					assets.NewVariableGroupReference("@(format_location(contact.fields.state)) Members"),
				},
			),
			`{
			"type": "add_contact_groups",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"groups": [
				{
					"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
					"name": "Testers"
				},
				{
					"name_match": "@(format_location(contact.fields.state)) Members"
				}
			]
		}`,
		},
		{
			actions.NewAddContactURNAction(
				actionUUID,
				"tel",
				"+234532626677",
			),
			`{
			"type": "add_contact_urn",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"scheme": "tel",
			"path": "+234532626677"
		}`,
		},
		{
			actions.NewAddInputLabelsAction(
				actionUUID,
				[]*assets.LabelReference{
					assets.NewLabelReference(assets.LabelUUID("3f65d88a-95dc-4140-9451-943e94e06fea"), "Spam"),
					assets.NewVariableLabelReference("@(format_location(contact.fields.state)) Messages"),
				},
			),
			`{
			"type": "add_input_labels",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"labels": [
				{
					"uuid": "3f65d88a-95dc-4140-9451-943e94e06fea",
					"name": "Spam"
				},
				{
					"name_match": "@(format_location(contact.fields.state)) Messages"
				}
			]
		}`,
		},
		{
			actions.NewCallResthookAction(
				actionUUID,
				"new-registration",
				"My Result",
			),
			`{
			"type": "call_resthook",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"resthook": "new-registration",
			"result_name": "My Result"
		}`,
		},
		{
			actions.NewCallWebhookAction(
				actionUUID,
				"POST",
				"http://example.com/ping",
				map[string]string{
					"Authentication": "Token @contact.fields.token",
				},
				`{"contact_id": 234}`, // body
				"Webhook Response",
			),
			`{
			"type": "call_webhook",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"method": "POST",
			"url": "http://example.com/ping",
			"headers": {
				"Authentication": "Token @contact.fields.token"
			},
			"body": "{\"contact_id\": 234}",
			"result_name": "Webhook Response"
		}`,
		},
		{
			actions.NewPlayAudioAction(
				actionUUID,
				"http://uploads.temba.io/2353262.m4a",
			),
			`{
			"type": "play_audio",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"audio_url": "http://uploads.temba.io/2353262.m4a"
		}`,
		},
		{
			actions.NewSayMsgAction(
				actionUUID,
				"Hi @contact.name, are you ready to complete today's survey?",
				"http://uploads.temba.io/2353262.m4a",
			),
			`{
			"type": "say_msg",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"audio_url": "http://uploads.temba.io/2353262.m4a",
			"text": "Hi @contact.name, are you ready to complete today's survey?"
		}`,
		},
		{
			actions.NewRemoveContactGroupsAction(
				actionUUID,
				[]*assets.GroupReference{
					assets.NewGroupReference(assets.GroupUUID("b7cf0d83-f1c9-411c-96fd-c511a4cfa86d"), "Testers"),
					assets.NewVariableGroupReference("@(format_location(contact.fields.state)) Members"),
				},
				false,
			),
			`{
			"type": "remove_contact_groups",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"groups": [
				{
					"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
					"name": "Testers"
				},
				{
					"name_match": "@(format_location(contact.fields.state)) Members"
				}
			]
		}`,
		},
		{
			actions.NewSendBroadcastAction(
				actionUUID,
				"Hi there",
				[]string{"http://example.com/red.jpg"},
				[]string{"Red", "Blue"},
				[]urns.URN{"twitter:nyaruka"},
				[]*flows.ContactReference{
					flows.NewContactReference(flows.ContactUUID("cbe87f5c-cda2-4f90-b5dd-0ac93a884950"), "Bob Smith"),
				},
				[]*assets.GroupReference{
					assets.NewGroupReference(assets.GroupUUID("b7cf0d83-f1c9-411c-96fd-c511a4cfa86d"), "Testers"),
				},
				nil,
			),
			`{
			"type": "send_broadcast",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"text": "Hi there",
			"attachments": ["http://example.com/red.jpg"],
			"quick_replies": ["Red", "Blue"],
			"urns": ["twitter:nyaruka"],
            "contacts": [
				{
					"uuid": "cbe87f5c-cda2-4f90-b5dd-0ac93a884950",
					"name": "Bob Smith"
				}
			],
			"groups": [
				{
					"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
					"name": "Testers"
				}
			]
		}`,
		},
		{
			actions.NewSendEmailAction(
				actionUUID,
				[]string{"bob@example.com"},
				"Hi there",
				"So I was thinking...",
			),
			`{
			"type": "send_email",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"addresses": ["bob@example.com"],
			"subject": "Hi there",
			"body": "So I was thinking..."
		}`,
		},
		{
			actions.NewSendMsgAction(
				actionUUID,
				"Hi there",
				[]string{"http://example.com/red.jpg"},
				[]string{"Red", "Blue"},
				true,
			),
			`{
			"type": "send_msg",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"text": "Hi there",
			"attachments": ["http://example.com/red.jpg"],
			"quick_replies": ["Red", "Blue"],
			"all_urns": true
		}`,
		},
		{
			actions.NewSetContactChannelAction(
				actionUUID,
				assets.NewChannelReference(assets.ChannelUUID("57f1078f-88aa-46f4-a59a-948a5739c03d"), "My Android Phone"),
			),
			`{
			"type": "set_contact_channel",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"channel": {
				"uuid": "57f1078f-88aa-46f4-a59a-948a5739c03d",
				"name": "My Android Phone"
			}
		}`,
		},
		{
			actions.NewSetContactFieldAction(
				actionUUID,
				assets.NewFieldReference("gender", "Gender"),
				"Male",
			),
			`{
			"type": "set_contact_field",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"field": {
				"key": "gender",
				"name": "Gender"
			},
			"value": "Male"
		}`,
		},
		{
			actions.NewSetContactLanguageAction(
				actionUUID,
				"eng",
			),
			`{
			"type": "set_contact_language",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"language": "eng"
		}`,
		},
		{
			actions.NewSetContactNameAction(
				actionUUID,
				"Bob",
			),
			`{
			"type": "set_contact_name",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"name": "Bob"
		}`,
		},
		{
			actions.NewSetContactTimezoneAction(
				actionUUID,
				"Africa/Kigali",
			),
			`{
			"type": "set_contact_timezone",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"timezone": "Africa/Kigali"
		}`,
		},
		{
			actions.NewSetRunResultAction(
				actionUUID,
				"Response 1",
				"yes",
				"Yes",
			),
			`{
			"type": "set_run_result",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"name": "Response 1",
			"value": "yes",
			"category": "Yes"
		}`,
		},
		{
			actions.NewEnterFlowAction(
				actionUUID,
				assets.NewFlowReference(assets.FlowUUID("fece6eac-9127-4343-9269-56e88f391562"), "Parent"),
				true, // terminal
			),
			`{
			"type": "enter_flow",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"flow": {
				"uuid": "fece6eac-9127-4343-9269-56e88f391562",
				"name": "Parent"
			},
			"terminal": true
		}`,
		},
		{
			actions.NewStartSessionAction(
				actionUUID,
				assets.NewFlowReference(assets.FlowUUID("fece6eac-9127-4343-9269-56e88f391562"), "Parent"),
				[]urns.URN{"twitter:nyaruka"},
				[]*flows.ContactReference{
					flows.NewContactReference(flows.ContactUUID("cbe87f5c-cda2-4f90-b5dd-0ac93a884950"), "Bob Smith"),
				},
				[]*assets.GroupReference{
					assets.NewGroupReference(assets.GroupUUID("b7cf0d83-f1c9-411c-96fd-c511a4cfa86d"), "Testers"),
				},
				nil,  // legacy vars
				true, // create new contact
			),
			`{
			"type": "start_session",
			"uuid": "ad154980-7bf7-4ab8-8728-545fd6378912",
			"flow": {
				"uuid": "fece6eac-9127-4343-9269-56e88f391562",
				"name": "Parent"
			},
			"urns": ["twitter:nyaruka"],
            "contacts": [
				{
					"uuid": "cbe87f5c-cda2-4f90-b5dd-0ac93a884950",
					"name": "Bob Smith"
				}
			],
			"groups": [
				{
					"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
					"name": "Testers"
				}
			],
			"create_contact": true
		}`,
		},
	}

	for _, tc := range tests {
		// test validating the action
		err := tc.action.Validate()
		assert.NoError(t, err)

		// test marshaling the action
		actualJSON, err := json.Marshal(tc.action)
		assert.NoError(t, err)

		test.AssertEqualJSON(t, json.RawMessage(tc.json), actualJSON, "new action produced unexpected JSON")
	}
}

func TestReadAction(t *testing.T) {
	// error if no type field
	_, err := actions.ReadAction([]byte(`{"foo": "bar"}`))
	assert.EqualError(t, err, "field 'type' is required")

	// error if we don't recognize action type
	_, err = actions.ReadAction([]byte(`{"type": "do_the_foo", "foo": "bar"}`))
	assert.EqualError(t, err, "unknown type: 'do_the_foo'")
}
