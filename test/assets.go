package test

import (
	"io/ioutil"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/assets/static"
	"github.com/nyaruka/goflow/assets/static/types"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/engine"
	"github.com/nyaruka/goflow/utils/uuids"
)

// LoadSessionAssets loads a session assets instance from a static JSON file
func LoadSessionAssets(path string) (flows.SessionAssets, error) {
	assetsJSON, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	source, err := static.NewSource(assetsJSON)
	if err != nil {
		return nil, err
	}

	return engine.NewSessionAssets(source, nil)
}

func LoadFlowFromAssets(path string, uuid assets.FlowUUID) (flows.Flow, error) {
	sa, err := LoadSessionAssets(path)
	if err != nil {
		return nil, err
	}

	return sa.Flows().Get(uuid)
}

func NewChannel(name string, address string, schemes []string, roles []assets.ChannelRole, parent *assets.ChannelReference) *flows.Channel {
	return flows.NewChannel(types.NewChannel(assets.ChannelUUID(uuids.New()), name, address, schemes, roles, parent))
}

func NewTelChannel(name string, address string, roles []assets.ChannelRole, parent *assets.ChannelReference, country string, matchPrefixes []string, allowInternational bool) *flows.Channel {
	return flows.NewChannel(types.NewTelChannel(assets.ChannelUUID(uuids.New()), name, address, roles, parent, country, matchPrefixes, allowInternational))
}

func NewClassifier(name, type_ string, intents []string) *flows.Classifier {
	return flows.NewClassifier(types.NewClassifier(assets.ClassifierUUID(uuids.New()), name, type_, intents))
}
