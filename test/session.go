package test

import (
	"encoding/json"
	"strings"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/assets/static"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/engine"
	"github.com/nyaruka/goflow/flows/resumes"
	"github.com/nyaruka/goflow/flows/triggers"

	"github.com/pkg/errors"
)

var sessionAssets = `{
    "channels": [
        {
            "uuid": "57f1078f-88aa-46f4-a59a-948a5739c03d",
            "name": "My Android Phone",
            "address": "+12345671111",
            "schemes": ["tel"],
            "roles": ["send", "receive"]
        },
        {
            "uuid": "8e21f093-99aa-413b-b55b-758b54308fcb",
            "name": "Twitter Channel",
            "address": "nyaruka",
            "schemes": ["twitter"],
            "roles": ["send", "receive"]
        },
        {
            "uuid": "4bb288a0-7fca-4da1-abe8-59a593aff648",
            "name": "Facebook Channel",
            "address": "235326346322111",
            "schemes": ["facebook"],
            "roles": ["send", "receive"]
        }
    ],
    "flows": [
        {
            "uuid": "50c3706e-fedb-42c0-8eab-dda3335714b7",
            "name": "Registration",
            "spec_version": "12.0",
            "language": "eng",
            "type": "messaging",
            "revision": 123,
            "nodes": [
                {
                    "uuid": "72a1f5df-49f9-45df-94c9-d86f7ea064e5",
                    "actions": [
                        {
                            "uuid": "9487a60e-a6ef-4a88-b35d-894bfe074144",
                            "type": "enter_flow",
                            "flow": {
                                "uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
                                "name": "Collect Age"
                            }
                        }
                    ],
                    "exits": [
                        {
                            "uuid": "d7a36118-0a38-4b35-a7e4-ae89042f0d3c",
                            "destination_node_uuid": "3dcccbb4-d29c-41dd-a01f-16d814c9ab82"
                        }
                    ]
                },
                {
                    "uuid": "3dcccbb4-d29c-41dd-a01f-16d814c9ab82",
                    "wait": {
                        "type": "msg",
                        "timeout": 600
                    },
                    "router": {
                        "type": "switch",
                        "default_exit_uuid": "37d8813f-1402-4ad2-9cc2-e9054a96525b",
                        "operand": "@input"
                    },
                    "exits": [
                        {
                            "uuid": "37d8813f-1402-4ad2-9cc2-e9054a96525b",
                            "name": "All Responses",
                            "destination_node_uuid": "f5bb9b7a-7b5e-45c3-8f0e-61b4e95edf03"
                        }
                    ]
                },
                {
                    "uuid": "f5bb9b7a-7b5e-45c3-8f0e-61b4e95edf03",
                    "actions": [
                        {
                            "uuid": "5508e6a7-26ce-4b3b-b32e-bb4e2e614f5d",
                            "type": "set_run_result",
                            "name": "Phone Number",
                            "value": "+12344563452"
                        },
                        {
                            "uuid": "72fea511-246f-49ad-846d-853b22ecc9c9",
                            "type": "set_run_result",
                            "name": "Favorite Color",
                            "value": "red",
                            "category": "Red"
                        },
                        {
                            "uuid": "821eef31-c6d2-45b1-8f6a-d396e4959bbf",
                            "type": "set_run_result",
                            "name": "2Factor",
                            "value": "34634624463525"
                        },
                        {
                            "uuid": "06153fbd-3e2c-413a-b0df-ed15d631835a",
                            "type": "call_webhook",
                            "method": "GET",
                            "url": "http://localhost/?content=%7B%22results%22%3A%5B%7B%22state%22%3A%22WA%22%7D%2C%7B%22state%22%3A%22IN%22%7D%5D%7D",
                            "result_name": "webhook"
                        }
                    ],
                    "exits": [
                        {
                            "uuid": "d898f9a4-f0fc-4ac4-a639-c98c602bb511",
                            "destination_node_uuid": "c0781400-737f-4940-9a6c-1ec1c3df0325"
                        }
                    ]
                },
                {
                    "uuid": "c0781400-737f-4940-9a6c-1ec1c3df0325",
                    "actions": [],
                    "exits": [
                        {
                            "uuid": "9fc5f8b4-2247-43db-b899-ab1ac50ba06c"
                        }
                    ]
                }
            ]
        },
        {
            "uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
            "name": "Collect Age",
            "spec_version": "12.0",
            "language": "eng",
            "type": "messaging",
            "nodes": [{
                "uuid": "d9dba561-b5ee-4f62-ba44-60c4dc242b84",
                "actions": [
                    {
                        "uuid": "4ed673b3-bdcc-40f2-944b-6ad1c82eb3ee",
                        "type": "set_run_result",
                        "name": "Age",
                        "value": "23",
                        "category": "Youth"
                    },
                    {
                        "uuid": "7a0c3cec-ef84-41aa-bf2b-be8259038683",
                        "type": "set_contact_field",
                        "field": {
                            "key": "age",
                            "name": "Age"
                        },
                        "value": "@results.age"
                    }
                ],
                "exits": [
                    {
                        "uuid": "4ee148c8-4026-41da-9d4c-08cb4d60b0d7"
                    }
                ]
            }]
        },
        {
            "uuid": "fece6eac-9127-4343-9269-56e88f391562",
            "name": "Parent",
            "spec_version": "12.0",
            "language": "eng",
            "type": "messaging",
            "nodes": []
        },
        {
            "uuid": "aa71426e-13bd-4607-a4f5-77666ff9c4bf",
            "name": "Voice Test",
            "spec_version": "12.0",
            "language": "eng",
            "type": "voice",
            "nodes": []
        }
    ],
    "fields": [
        {"key": "gender", "name": "Gender", "type": "text"},
        {"key": "age", "name": "Age", "type": "number"},
        {"key": "join_date", "name": "Join Date", "type": "datetime"},
        {"key": "activation_token", "name": "Activation Token", "type": "text"},
        {"key": "not_set", "name": "Not set", "type": "text"}
    ],
    "groups": [
        {"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d", "name": "Testers"},
        {"uuid": "4f1f98fc-27a7-4a69-bbdb-24744ba739a9", "name": "Males"},
        {"uuid": "1e1ce1e1-9288-4504-869e-022d1003c72a", "name": "Customers"}
    ],
    "labels": [
        {"uuid": "3f65d88a-95dc-4140-9451-943e94e06fea", "name": "Spam"}
    ],
    "locations": [
        {
            "name": "Rwanda",
            "aliases": ["Ruanda"],		
            "children": [
                {
                    "name": "Kigali City",
                    "aliases": ["Kigali", "Kigari"],
                    "children": [
                        {
                            "name": "Gasabo",
                            "children": [
                                {
                                    "name": "Gisozi"
                                },
                                {
                                    "name": "Ndera"
                                }
                            ]
                        },
                        {
                            "name": "Nyarugenge",
                            "children": []
                        }
                    ]
                }
            ]
        }
    ],
    "resthooks": [
        {
            "slug": "new-registration", 
            "subscribers": [
                "http://localhost/?cmd=success"
            ]
        }
    ]
}`

var sessionTrigger = `{
    "type": "flow_action",
    "triggered_on": "2017-12-31T11:31:15.035757258-02:00",
    "flow": {"uuid": "50c3706e-fedb-42c0-8eab-dda3335714b7", "name": "Registration"},
    "contact": {
        "uuid": "5d76d86b-3bb9-4d5a-b822-c9d86f5d8e4f",
        "id": 1234567,
        "name": "Ryan Lewis",
        "language": "eng",
        "timezone": "America/Guayaquil",
        "created_on": "2018-06-20T11:40:30.123456789-00:00",
        "urns": [
            "tel:+12065551212?channel=57f1078f-88aa-46f4-a59a-948a5739c03d", 
            "twitterid:54784326227#nyaruka",
            "mailto:foo@bar.com"
        ],
        "groups": [
            {"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d", "name": "Testers"},
            {"uuid": "4f1f98fc-27a7-4a69-bbdb-24744ba739a9", "name": "Males"}
        ],
        "fields": {
            "gender": {
                "text": "Male"
            },
            "join_date": {
                "text": "2017-12-02", "datetime": "2017-12-02T00:00:00-02:00"
            },
            "activation_token": {
                "text": "AACC55"
            }
        }
    },
    "run_summary": {
        "uuid": "4213ac47-93fd-48c4-af12-7da8218ef09d",
        "contact": {
            "uuid": "c59b0033-e748-4240-9d4c-e85eb6800151",
            "name": "Jasmine",
            "created_on": "2018-01-01T12:00:00.000000000-00:00",
            "language": "spa",
            "urns": [
                "tel:+593979111222"
            ],
            "fields": {
                "age": {
                    "text": "33 years", "number": 33
                },
                "gender": {
                    "text": "Female"
                }
            }
        },
        "flow": {
            "uuid": "fece6eac-9127-4343-9269-56e88f391562",
            "name": "Parent Flow"
        },
        "results": {
            "role": {
                "created_on": "2000-01-01T00:00:00.000000000-00:00",
                "input": "a reporter",
                "name": "Role",
                "node_uuid": "385cb848-5043-448e-9123-05cbcf26ad74",
                "value": "reporter",
                "category": "Reporter"
            }
        },
        "status": "active"
    },
    "environment": {
        "date_format": "YYYY-MM-DD",
        "default_language": "eng",
        "allowed_languages": [
            "eng", 
            "spa"
        ],
        "redaction_policy": "none",
        "time_format": "hh:mm",
        "timezone": "America/Guayaquil"
    },
    "params": {"source": "website","address": {"state": "WA"}}
}`

var sessionResume = `{
    "type": "msg",
    "msg": {
        "attachments": [
            "image/jpeg:http://s3.amazon.com/bucket/test.jpg",
            "audio/mp3:http://s3.amazon.com/bucket/test.mp3"
        ],
        "channel": {
            "name": "Nexmo",
            "uuid": "57f1078f-88aa-46f4-a59a-948a5739c03d"
        },
        "text": "Hi there",
        "urn": "tel:+12065551212",
        "uuid": "9bf91c2b-ce58-4cef-aacc-281e03f69ab5"
    },
    "resumed_on": "2017-12-31T11:35:10.035757258-02:00"
}`

var voiceSessionAssets = `{
    "channels": [
        {
            "uuid": "57f1078f-88aa-46f4-a59a-948a5739c03d",
            "name": "My Android Phone",
            "address": "+12345671111",
            "schemes": ["tel"],
            "roles": ["send", "receive"]
        },
        {
            "uuid": "fd47a886-451b-46fb-bcb6-242a4046c0c0",
            "name": "Nexmo",
            "address": "345642627",
            "schemes": ["tel"],
            "roles": ["send", "receive", "call", "answer"]
        }
    ],
    "flows": [
        {
            "uuid": "aa71426e-13bd-4607-a4f5-77666ff9c4bf",
            "name": "Voice Test",
            "spec_version": "12.0",
            "language": "eng",
            "type": "voice",
            "nodes": [
                {
                    "uuid": "6da04a32-6c84-40d9-b614-3782fde7af80",
                    "type": "set_run_result",
                    "name": "Age",
                    "value": "23",
                    "category": "Youth"
                }
            ]
        }
    ],
    "fields": [
        {"key": "gender", "name": "Gender", "type": "text"}
    ],
    "groups": [
        {"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d", "name": "Testers"},
        {"uuid": "4f1f98fc-27a7-4a69-bbdb-24744ba739a9", "name": "Males"},
        {"uuid": "1e1ce1e1-9288-4504-869e-022d1003c72a", "name": "Customers"}
    ]
}`

var voiceSessionTrigger = `{
    "type": "channel",
    "triggered_on": "2017-12-31T11:31:15.035757258-02:00",
    "event": {
        "type": "incoming_call",
        "channel": {"uuid": "fd47a886-451b-46fb-bcb6-242a4046c0c0", "name": "Nexmo"}
    },
    "connection": {
        "channel": {"uuid": "fd47a886-451b-46fb-bcb6-242a4046c0c0", "name": "Nexmo"},
        "urn": "tel:+12065551212"
    },
    "flow": {"uuid": "aa71426e-13bd-4607-a4f5-77666ff9c4bf", "name": "Voice Test"},
    "contact": {
        "uuid": "5d76d86b-3bb9-4d5a-b822-c9d86f5d8e4f",
        "id": 1234567,
        "name": "Ryan Lewis",
        "language": "eng",
        "timezone": "America/Guayaquil",
        "created_on": "2018-06-20T11:40:30.123456789-00:00",
        "urns": [
            "tel:+12065551212"
        ],
        "groups": [
            {"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d", "name": "Testers"},
            {"uuid": "4f1f98fc-27a7-4a69-bbdb-24744ba739a9", "name": "Males"}
        ],
        "fields": {
            "gender": {
                "text": "Male"
            }
        }
    },
    "environment": {
        "date_format": "DD-MM-YYYY",
        "default_language": "eng",
        "allowed_languages": [
            "eng", 
            "spa"
        ],
        "redaction_policy": "none",
        "time_format": "hh:mm",
        "timezone": "America/Guayaquil"
    }
}`

// CreateTestSession creates a standard example session for testing
func CreateTestSession(testServerURL string, actionToAdd flows.Action) (flows.Session, []flows.Event, error) {

	session, err := CreateSession(json.RawMessage(sessionAssets), testServerURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating test session")
	}

	// optional modify the main flow by adding the provided action to the last node
	if actionToAdd != nil {
		flow, _ := session.Assets().Flows().Get(assets.FlowUUID("50c3706e-fedb-42c0-8eab-dda3335714b7"))
		flow.Nodes()[len(flow.Nodes())-1].AddAction(actionToAdd)
	}

	// read our trigger
	trigger, err := triggers.ReadTrigger(session.Assets(), json.RawMessage(sessionTrigger), assets.PanicOnMissing)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error reading trigger")
	}

	_, err = session.Start(trigger)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error starting test session")
	}

	// read our resume
	resume, err := resumes.ReadResume(session.Assets(), json.RawMessage(sessionResume), assets.PanicOnMissing)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error reading resume")
	}

	sprint, err := session.Resume(resume)
	return session, sprint.Events(), err
}

// CreateTestVoiceSession creates a standard example session for testing voice flows and actions
func CreateTestVoiceSession(testServerURL string, actionToAdd flows.Action) (flows.Session, []flows.Event, error) {

	session, err := CreateSession(json.RawMessage(voiceSessionAssets), testServerURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating test voice session")
	}

	// optional modify the main flow by adding the provided action to the last node
	if actionToAdd != nil {
		flow, _ := session.Assets().Flows().Get(assets.FlowUUID("aa71426e-13bd-4607-a4f5-77666ff9c4bf"))
		nodes := flow.Nodes()
		nodes[len(nodes)-1].AddAction(actionToAdd)
	}

	// read our trigger
	trigger, err := triggers.ReadTrigger(session.Assets(), json.RawMessage(voiceSessionTrigger), assets.PanicOnMissing)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error reading trigger")
	}

	sprint, err := session.Start(trigger)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error starting test voice session")
	}

	return session, sprint.Events(), err
}

// CreateSession creates a session with the given assets
func CreateSession(assetsJSON json.RawMessage, testServerURL string) (flows.Session, error) {
	// different tests different ports for the test HTTP server
	if testServerURL != "" {
		assetsJSON = json.RawMessage(strings.Replace(string(assetsJSON), "http://localhost", testServerURL, -1))
	}

	// read our assets into a source
	source, err := static.NewSource(assetsJSON)
	if err != nil {
		return nil, errors.Wrap(err, "error loading test assets")
	}

	// create our engine session
	assets, err := engine.NewSessionAssets(source)
	if err != nil {
		return nil, errors.Wrap(err, "error creating test session assets")
	}

	eng := engine.NewBuilder().WithDefaultUserAgent("goflow-testing").Build()
	session := eng.NewSession(assets)
	return session, nil
}
