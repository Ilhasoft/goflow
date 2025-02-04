package inputs

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/utils"

	"github.com/pkg/errors"
)

type readFunc func(flows.SessionAssets, json.RawMessage, assets.MissingCallback) (flows.Input, error)

var registeredTypes = map[string]readFunc{}

// RegisterType registers a new type of input
func RegisterType(name string, f readFunc) {
	registeredTypes[name] = f
}

type baseInput struct {
	type_     string
	uuid      flows.InputUUID
	channel   *flows.Channel
	createdOn time.Time
}

func newBaseInput(typeName string, uuid flows.InputUUID, channel *flows.Channel, createdOn time.Time) baseInput {
	return baseInput{
		type_:     typeName,
		uuid:      uuid,
		channel:   channel,
		createdOn: createdOn,
	}
}

// Type returns the type of this input
func (i *baseInput) Type() string { return i.type_ }

func (i *baseInput) UUID() flows.InputUUID   { return i.uuid }
func (i *baseInput) Channel() *flows.Channel { return i.channel }
func (i *baseInput) CreatedOn() time.Time    { return i.createdOn }

// Resolve resolves the given key when this input is referenced in an expression
func (i *baseInput) Resolve(env utils.Environment, key string) types.XValue {
	switch strings.ToLower(key) {
	case "type":
		return types.NewXText(i.type_)
	case "uuid":
		return types.NewXText(string(i.uuid))
	case "created_on":
		return types.NewXDateTime(i.createdOn)
	case "channel":
		return i.channel
	}

	return types.NewXResolveError(i, key)
}

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

type baseInputEnvelope struct {
	Type      string                   `json:"type" validate:"required"`
	UUID      flows.InputUUID          `json:"uuid"`
	Channel   *assets.ChannelReference `json:"channel,omitempty" validate:"omitempty,dive"`
	CreatedOn time.Time                `json:"created_on" validate:"required"`
}

// ReadInput reads an input from the given typed envelope
func ReadInput(sessionAssets flows.SessionAssets, data json.RawMessage, missing assets.MissingCallback) (flows.Input, error) {
	typeName, err := utils.ReadTypeFromJSON(data)
	if err != nil {
		return nil, err
	}

	f := registeredTypes[typeName]
	if f == nil {
		return nil, errors.Errorf("unknown type: '%s'", typeName)
	}

	return f(sessionAssets, data, missing)
}

func (i *baseInput) unmarshal(sessionAssets flows.SessionAssets, e *baseInputEnvelope, missing assets.MissingCallback) error {
	i.type_ = e.Type
	i.uuid = e.UUID
	i.createdOn = e.CreatedOn

	if e.Channel != nil {
		i.channel = sessionAssets.Channels().Get(e.Channel.UUID)
		if i.channel == nil {
			missing(e.Channel)
			return nil
		}
	}
	return nil
}

func (i *baseInput) marshal(e *baseInputEnvelope) {
	e.Type = i.type_
	e.UUID = i.uuid
	e.CreatedOn = i.createdOn
	e.Channel = i.channel.Reference()
}
