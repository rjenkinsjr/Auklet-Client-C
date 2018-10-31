package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/vmihailenco/msgpack"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// Converter converts a stream of agent.Message to a stream of broker.Message.
type Converter struct {
	in  MessageSource
	out chan broker.Message
	Config
}

// ExitSignalApp is an App that has a signal and exit status.
type ExitSignalApp interface {
	AgentVersion() string
	CheckSum() string
	ExitStatus() int
	Signal() string
}

// MessageSource is a source of agent messages.
type MessageSource interface {
	Output() <-chan agent.Message
}

// Persistor provides a persistor interface.
type Persistor interface {
	CreateMessage(*broker.Message) error
}

// Monitor provides system metrics.
type Monitor interface {
	GetMetrics() device.Metrics
	Close()
}

// Config provides parameters needed by a Converter.
type Config struct {
	Monitor     Monitor
	Persistor   Persistor
	App         ExitSignalApp
	Username    string
	UserVersion string
	AppID       string
	MacHash     string
	Encoding    Encoding
}

// Encoding represents the serialization encoding.
type Encoding int

const (
	MsgPack Encoding = iota
	JSON
)

// NewConverter returns a converter for the given input streams that uses the
// given persistor and app.
func NewConverter(cfg Config, in ...agent.MessageSource) Converter {
	c := Converter{
		in:     agent.Merge(in...),
		out:    make(chan broker.Message),
		Config: cfg,
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
	defer c.Monitor.Close()
	for agentMsg := range c.in.Output() {
		switch agentMsg.Type {
		case "applog", "log":
			// Drop these messages for now, because consumers do not handle them.
			continue
		}

		brokerMsg := c.convert(agentMsg)
		if c.Persistor != nil {
			if err := c.Persistor.CreateMessage(&brokerMsg); err != nil {
				// Let the backend know we ran out of local storage.
				c.out <- broker.Message{
					Error: err.Error(),
					Topic: broker.Log,
				}
				continue
			}
		}

		c.out <- brokerMsg
	}
}

func (c Converter) convert(m agent.Message) broker.Message {
	switch m.Type {
	case "applog":
		return c.marshal(c.appLog(m.Data), broker.Event)
	case "profile":
		return c.marshal(c.profile(m.Data), broker.Profile)
	case "event":
		log.Printf("%v exited with error signal", c.App)
		return c.marshal(c.errorSig(m.Data), broker.Event)
	case "log":
		return broker.Message{
			Bytes: m.Data,
			Topic: broker.Log,
		}
	case "cleanExit":
		// This message is not actually generated by the agent, but
		// by agent.Server, if it receives EOF and has not seen an
		// "event".
		return c.marshal(c.exit(), broker.Event)
	default:
		return broker.Message{
			Error: fmt.Sprintf("message of type %q not handled", m.Type),
			Topic: broker.Log,
		}
	}
}

// msgpackMarshal has the same signature as json.Marshal, so that the two
// functions can be interchanged.
func msgpackMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseJSONTag(true)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

func (c Converter) marshal(v interface{}, topic broker.Topic) broker.Message {
	marshaler := map[Encoding]func(interface{}) ([]byte, error){
		MsgPack: msgpackMarshal,
		JSON:    json.Marshal,
	}[c.Encoding]
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
