// Package agent provides access to an Auklet agent.
package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

// Message represents messages that can be received by a Server, and thus,
// would be sent by an agent.
type Message struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
	Error string
}

// Server provides a connection server for an Auklet agent.
type Server struct {
	in  io.Reader
	dec *json.Decoder
	out chan Message
	// Done closes when the Server gets EOF.
	Done chan struct{}
}

// NewServer returns a new Server that reads from in. If dec is not nil, it is
// used directly.
func NewServer(in io.Reader, dec *json.Decoder) *Server {
	s := &Server{
		in:   in,
		dec:  dec,
		out:  make(chan Message),
		Done: make(chan struct{}),
	}
	go s.serve()
	return s
}

// serve causes s to accept an incoming connection, after which s can send and
// receive messages.
func (s *Server) serve() {
	defer close(s.Done)
	defer close(s.out)
	log.Print("Server: accepted connection")
	defer log.Print("Server: connection closed")
	if s.dec == nil {
		s.dec = json.NewDecoder(s.in)
	}
	errd := false
	for {
		var msg Message
		if err := s.dec.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			// There was a problem decoding the stream into
			// message format.
			buf, _ := ioutil.ReadAll(s.dec.Buffered())
			s.out <- Message{
				Type:  "log",
				Error: fmt.Sprintf("%v in %v", err.Error(), string(buf)),
			}
			s.dec = json.NewDecoder(s.in)
			continue
		}
		if msg.Type == "event" {
			errd = true
		}
		s.out <- msg
	}
	if !errd {
		s.out <- Message{Type: "cleanExit"}
	}
}

// Output returns s's output stream.
func (s *Server) Output() <-chan Message {
	return s.out
}
