package message

import (
	"errors"
	"testing"
	"time"

	"github.com/aukletio/Auklet-Client-C/api"
	"github.com/aukletio/Auklet-Client-C/broker"
)

func TestEnsureFuture(t *testing.T) {
	cases := []struct {
		day         int
		now, expect time.Time
	}{
		{
			day:    1,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expect: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			day:    1,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 1, time.UTC),
			expect: time.Date(2000, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			day:    12,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expect: time.Date(2000, 1, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			day:    15,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expect: time.Date(2000, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for i, c := range cases {
		if d := ensureFuture(dayThisMonth(c.day, c.now), c.now); !c.expect.Equal(d) {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, d)
		}
	}
}

func expiredTimer() *time.Timer {
	c := make(chan time.Time)
	close(c)
	return &time.Timer{C: c}
}

func closedChan() <-chan broker.Message {
	c := make(chan broker.Message)
	close(c)
	return c
}

func sendOne() <-chan broker.Message {
	c := make(chan broker.Message)
	go func() { c <- broker.Message{} }()
	return c
}

func sendConf() chan api.CellularConfig {
	c := make(chan api.CellularConfig)
	go func() { c <- api.CellularConfig{} }()
	return c
}

func TestStateFuncs(t *testing.T) {
	cases := []struct {
		state  state // which state to test
		l      *DataLimiter
		expect state
	}{
		{
			state:  initial,
			l:      &DataLimiter{},
			expect: underBudget,
		},
		{
			state: initial,
			l: &DataLimiter{
				Budget:    10,
				HasBudget: true,
				Count:     10,
			},
			expect: overBudget,
		},
		{
			state: underBudget,
			l: &DataLimiter{
				periodTimer: expiredTimer(),
				store:       new(MemPersistor),
			},
			expect: initial,
		},
		{
			state:  underBudget,
			l:      &DataLimiter{periodTimer: new(time.Timer), in: closedChan()},
			expect: cleanup,
		},
		{
			state: underBudget,
			l: &DataLimiter{
				periodTimer: new(time.Timer),
				in:          sendOne(),
				out:         make(chan broker.Message, 1),
				store:       new(MemPersistor),
			},
			expect: underBudget,
		},
		{
			state: underBudget,
			l: &DataLimiter{
				periodTimer: new(time.Timer),
				store:       new(MemPersistor),
				conf:        sendConf(),
			},
			expect: initial,
		},
		{
			state: overBudget,
			l: &DataLimiter{
				periodTimer: expiredTimer(),
				store:       new(MemPersistor),
			},
			expect: initial,
		},
		{
			state:  overBudget,
			l:      &DataLimiter{in: closedChan(), periodTimer: new(time.Timer)},
			expect: cleanup,
		},
		{
			state: overBudget,
			l: &DataLimiter{
				in:          sendOne(),
				periodTimer: new(time.Timer),
			},
			expect: overBudget,
		},
		{
			state: overBudget,
			l: &DataLimiter{
				conf:        sendConf(),
				store:       new(MemPersistor),
				periodTimer: new(time.Timer),
			},
			expect: initial,
		},
		{
			state:  cleanup,
			l:      &DataLimiter{out: make(chan broker.Message)},
			expect: terminal,
		},
	}

	for i, c := range cases {
		if got := c.l.lookup(c.state)(); got != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, got)
		}
	}
}

func (s state) String() string {
	return map[state]string{
		terminal:    "terminal",
		initial:     "initial",
		underBudget: "underBudget",
		overBudget:  "overBudget",
		cleanup:     "cleanup",
	}[s]
}

func TestHandleMessage(t *testing.T) {
	cases := []struct {
		l      *DataLimiter
		m      broker.Message
		expect state
	}{
		{
			l:      &DataLimiter{Budget: 10, HasBudget: true},
			m:      broker.Message{Bytes: make([]byte, 100)},
			expect: overBudget,
		},
		{
			l: &DataLimiter{
				Count:     85,
				Budget:    100,
				HasBudget: true,
				out:       make(chan broker.Message, 1),
				store:     new(MemPersistor),
			},
			m:      broker.Message{Bytes: make([]byte, 10)},
			expect: overBudget,
		},
		{
			l: &DataLimiter{
				Budget:    100,
				HasBudget: true,
				out:       make(chan broker.Message, 1),
				store:     mockPers{},
			},
			expect: overBudget,
		},
		{
			l: &DataLimiter{
				out:   make(chan broker.Message, 1),
				store: new(MemPersistor),
			},
			expect: underBudget,
		},
	}

	for i, c := range cases {
		if got := c.l.handleMessage(c.m); got != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, got)
		}
	}
}

type mockPers struct{}

var errPers = errors.New("mock error")

func (mockPers) Save(Encodable) error { return errPers }
func (mockPers) Load(Decodable) error { return errPers }

func TestSetBudget(t *testing.T) {
	cases := []struct {
		mb        int
		hasBudget bool
		expect    int
	}{
		{},
		{mb: 1, hasBudget: true, expect: 1000000},
	}

	for i, c := range cases {
		l := &DataLimiter{}
		if l.setBudget(c.mb, c.hasBudget); l.Budget != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, l.Budget)
		}
	}
}
