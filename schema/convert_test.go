package schema

import (
	"testing"

	"github.com/aukletio/Auklet-Client-C/agent"
	"github.com/aukletio/Auklet-Client-C/broker"
	"github.com/aukletio/Auklet-Client-C/device"
)

type persistor struct{}

func (p persistor) CreateMessage(*broker.Message) error { return nil }

type source chan agent.Message

func (s source) Output() <-chan agent.Message { return s }

type app struct{}

func (app) CheckSum() string     { return "checksum" }
func (app) ExitStatus() int      { return 42 }
func (app) Signal() string       { return "something" }
func (app) AgentVersion() string { return "something" }

type monitor struct{}

func (monitor) GetMetrics() device.Metrics { return device.Metrics{} }
func (monitor) Close()                     {}

var cfg = Config{
	Monitor:     monitor{},
	Persistor:   persistor{},
	App:         app{},
	Username:    "username",
	UserVersion: "userVersion",
	AppID:       "app id",
	MacHash:     "mac hash",
}

func TestConverter(t *testing.T) {
	type converterCase struct {
		input agent.Message
		ok    bool
	}
	cases := []converterCase{
		{input: agent.Message{Type: "event"}, ok: true},
		{input: agent.Message{Type: "profile"}, ok: true},
		{input: agent.Message{Type: "cleanExit"}, ok: true},
		{input: agent.Message{Type: "unknown"}, ok: false},
	}
	for i, c := range cases {
		s := make(source)
		converter := NewConverter(cfg, s)
		s <- c.input
		m := <-converter.Output()
		ok := m.Error == ""
		if ok != c.ok {
			t.Errorf("case %v: got %v, expected %v: %v", i, ok, c.ok, m.Error)
		}
		close(s)
	}
}
