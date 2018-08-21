package schema

import (
	"bytes"
	"fmt"
	"log"

	"github.com/vmihailenco/msgpack"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// Converter converts a stream of agent.Message to a stream of broker.Message.
type Converter struct {
	in        MessageSource
	out       chan broker.Message
	persistor Persistor
	app       ExitWaitSignalApp
	username  string
}

// ExitWaitSignalApp is an ExitApp for which we can wait to exit.
type ExitWaitSignalApp interface {
	AgentVersion() string
	CheckSum() string
	ExitStatus() int
	Signal() string
	Wait()
}

// MessageSource is a source of agent messages.
type MessageSource interface {
	Output() <-chan agent.Message
}

// Persistor provides a persistor interface.
type Persistor interface {
	CreateMessage(*broker.Message) error
}

// NewConverter returns a converter for the given input stream that uses the
// given persistor and app.
func NewConverter(in MessageSource, persistor Persistor, app ExitWaitSignalApp, username string) Converter {
	c := Converter{
		in:        in,
		out:       make(chan broker.Message),
		persistor: persistor,
		app:       app,
		username:  username,
	}
	go c.serve()
	return c
}

// Output returns the converter's output stream.
func (c Converter) Output() <-chan broker.Message {
	return c.out
}

func (c Converter) serve() {
	defer close(c.out)
	for agentMsg := range c.in.Output() {
		brokerMsg := c.convert(agentMsg)
		if err := c.persistor.CreateMessage(&brokerMsg); err != nil {
			// Let the backend know we ran out of local storage.
			c.out <- broker.Message{
				Error: err.Error(),
				Topic: broker.Log,
			}
			continue
		}
		c.out <- brokerMsg
	}
}

func (c Converter) convert(m agent.Message) broker.Message {
	switch m.Type {
	case "applog":
		return marshal(c.appLog(m.Data), broker.Event)
	case "profile":
		return marshal(c.profile(m.Data), broker.Profile)
	case "event":
		c.app.Wait()
		log.Printf("%v exited with error signal", c.app)
		return marshal(c.errorSig(m.Data), broker.Event)
	case "log":
		return broker.Message{
			Bytes: m.Data,
			Topic: broker.Log,
		}
	case "cleanExit":
		// This message is not actually generated by the agent, but
		// by agent.Server, if it receives EOF and has not seen an
		// "event".
		return marshal(c.exit(), broker.Event)
	default:
		return broker.Message{
			Error: fmt.Sprintf("message of type %q not handled", m.Type),
			Topic: broker.Log,
		}
	}
}

// marshaler determines which transport encoding is used for messages.
var marshaler = msgpackMarshal

// msgpackMarshal has the same signature as json.Marshal, so that the two
// functions can be interchanged.
func msgpackMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseJSONTag(true)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

func marshal(v interface{}, topic broker.Topic) broker.Message {
	bytes, err := marshaler(v)
	return broker.Message{
		Error: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
		Bytes: bytes,
		Topic: topic,
	}
}
