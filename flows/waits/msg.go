package waits

import (
	"encoding/json"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/flows/resumes"
	"github.com/nyaruka/goflow/flows/triggers"
	"github.com/nyaruka/goflow/flows/waits/hints"
	"github.com/nyaruka/goflow/utils"

	"github.com/pkg/errors"
)

func init() {
	RegisterType(TypeMsg, readMsgWait)
}

// TypeMsg is the type of our message wait
const TypeMsg string = "msg"

// MsgWait is a wait which waits for an incoming message (i.e. a msg_received event)
type MsgWait struct {
	baseWait

	// Message waits can indicate to the caller what kind of message the flow is expecting. In the case of flows of type
	// messaging_offline, this should be considered a requirement and the client should only reply with a message containing
	// an attachment of that type. In the case of other flow types this should be considered only a hint to the channel,
	// which may or may not support prompting the contact for media of that type.
	hint flows.Hint
}

// NewMsgWait creates a new message wait
func NewMsgWait(timeout *int, hint flows.Hint) *MsgWait {
	return &MsgWait{
		baseWait: newBaseWait(TypeMsg, timeout),
		hint:     hint,
	}
}

// Hint returns the hint (optional)
func (w *MsgWait) Hint() flows.Hint { return w.hint }

// Begin beings waiting at this wait
func (w *MsgWait) Begin(run flows.FlowRun, log flows.EventCallback) bool {
	if !w.baseWait.Begin(run) {
		return false
	}

	// if we have a msg trigger and we're the first thing to happen... then we skip ourselves
	triggerHasMsg := run.Session().Trigger().Type() == triggers.TypeMsg

	if triggerHasMsg && len(run.Session().Runs()) == 1 && len(run.Path()) == 1 {
		return false
	}

	log(events.NewMsgWait(w.timeoutOn))
	return true
}

// End ends this wait or returns an error
func (w *MsgWait) End(resume flows.Resume, node flows.Node) error {
	// if we have a message we can definitely resume
	if resume.Type() == resumes.TypeMsg {
		return nil
	}

	return w.baseWait.End(resume, node)
}

var _ flows.Wait = (*MsgWait)(nil)

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

type msgWaitEnvelope struct {
	baseWaitEnvelope

	Hint json.RawMessage `json:"hint,omitempty"`
}

func readMsgWait(data json.RawMessage) (flows.Wait, error) {
	e := &msgWaitEnvelope{}
	if err := utils.UnmarshalAndValidate(data, e); err != nil {
		return nil, err
	}

	w := &MsgWait{}

	var err error
	if e.Hint != nil {
		if w.hint, err = hints.ReadHint(e.Hint); err != nil {
			return nil, errors.Wrap(err, "unable to read hint")
		}
	}

	return w, w.unmarshal(&e.baseWaitEnvelope)
}

// MarshalJSON marshals this wait into JSON
func (w *MsgWait) MarshalJSON() ([]byte, error) {
	e := &msgWaitEnvelope{}

	if err := w.marshal(&e.baseWaitEnvelope); err != nil {
		return nil, err
	}

	var err error
	if w.hint != nil {
		if e.Hint, err = json.Marshal(w.hint); err != nil {
			return nil, err
		}
	}

	return json.Marshal(e)
}
