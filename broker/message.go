package broker

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// This file defines interfaces for manipulating streams of broker
// messages, plus a message persistence layer.

// Topic encodes a Message topic.
type Topic int

// Profile, Event, and Log are Message types.
const (
	Profile Topic = iota
	Event
	Log
)

// Message represents a broker message.
type Message struct {
	// Type encodes the concrete Go type, for use by adapters.
	Type  string          `json:"type"`
	Topic Topic           `json:"topic"`
	Bytes json.RawMessage `json:"bytes"`
	path  string
}

// ErrStorageFull indicates that the corresponding Persistor is full.
type ErrStorageFull struct {
	limit int64
	count int
}

// Error returns e as a string.
func (e ErrStorageFull) Error() string {
	return fmt.Sprintf("persistor: storage full: %v used of %v limit", e.count, e.limit)
}

// Persistor controls a persistence layer for Messages.
type Persistor struct {
	limit        *int64      // storage limit in bytes; no limit if nil
	newLimit     chan *int64 // incoming new values for limit
	currentLimit chan *int64 // outgoing current values for limit
	dir          string
	count        int // counter to give Messages unique names
	out          chan Message
}

// StdPersistor is the standard Persistor.
var StdPersistor = NewPersistor(".auklet/message")

// CreateMessage creates a new Message under the standard Persistor. v is
// assumed to be losslessly encodable via the encoding/json package.
func CreateMessage(v interface{}, topic Topic) (m Message, err error) {
	return StdPersistor.CreateMessage(v, topic)
}

// NewPersistor creates a new Persistor in dir.
func NewPersistor(dir string) *Persistor {
	if err := os.MkdirAll(dir, 0777); err != nil {
		errorlog.Printf("persistor: unable to save unsent messages to %v: %v", dir, err)
	}
	p := &Persistor{
		dir:          dir,
		newLimit:     make(chan *int64),
		currentLimit: make(chan *int64),
	}
	p.load()
	go p.serve()
	return p
}

// serve serializes access to p.limit
func (p *Persistor) serve() {
	for {
		select {
		case p.limit = <-p.newLimit:
		case p.currentLimit <- p.limit:
		}
	}
}

// Configure returns a channel on which p's storage limit can be controlled.
func (p *Persistor) Configure() chan<- *int64 {
	return p.newLimit
}

// filepaths returns a list of paths of persistent messages.
func (p *Persistor) filepaths() (paths []string) {
	d, err := os.Open(p.dir)
	if err != nil {
		errorlog.Printf("persistor: failed to open message directory: %v", err)
		return
	}
	defer d.Close()
	names, err := d.Readdirnames(0)
	if err != nil {
		errorlog.Printf("persistor: failed to read directory names in %v: %v", d.Name(), err)
		return
	}
	for _, name := range names {
		paths = append(paths, p.dir+"/"+name)
	}
	return
}

func (p *Persistor) size() (n int64) {
	for _, path := range p.filepaths() {
		f, err := os.Stat(path)
		if err != nil {
			errorlog.Printf("persistor: failed to calculate storage size of message %v: %v", path, err)
			continue
		}
		n += f.Size()
	}
	return
}

// load loads the output channel with messages from the filesystem.
func (p *Persistor) load() {
	paths := p.filepaths()
	p.out = make(chan Message, len(paths))
	defer close(p.out)
	for _, path := range paths {
		m := Message{path: path}
		if m.load() != nil {
			continue
		}
		p.out <- m
	}
}

// Output returns p's output channel, which closes after all persisted messages
// have been sent.
func (p *Persistor) Output() <-chan Message {
	return p.out
}

// CreateMessage creates a new Message under p.
func (p *Persistor) CreateMessage(v interface{}, topic Topic) (m Message, err error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return
	}
	lim := <-p.currentLimit
	if lim != nil && int64(len(bytes))+p.size() > 9**lim/10 {
		err = ErrStorageFull{
			limit: *lim,
			count: p.count,
		}
		return
	}
	m = Message{
		Type:  fmt.Sprintf("%T", v),
		Topic: topic,
		Bytes: bytes,
		path:  fmt.Sprintf("%v/%v-%v", p.dir, os.Getpid(), p.count),
	}
	p.count++
	// Failing to save a message is a recoverable error that does not affect
	// our caller's logic. Thus, we don't return save's error value.
	m.save()
	return
}

func (m Message) load() (err error) {
	defer func() {
		if err != nil {
			errorlog.Printf("persistor: failed to load message %v: %v", m.path, err)
		}
	}()
	f, err := os.Open(m.path)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&m)
	return
}

func (m Message) save() (err error) {
	defer func() {
		if err != nil {
			errorlog.Printf("persistor: failed to save message %v: %v", m.path, err)
		}
	}()
	f, err := os.OpenFile(m.path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(m)
	return
}

// Remove deletes m from the persistence layer.
func (m Message) Remove() {
	os.Remove(m.path)
}

// MessageSource is implemented by types that can generate a Message stream.
type MessageSource interface {
	// Output returns a channel of Messages provided by a Source. A source
	// indicates when it has no more Messages to send by closing the
	// channel.
	Output() <-chan Message
}
