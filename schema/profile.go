package schema

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// profile represents profile data as expected by broker consumers.
type profile struct {
	// AppID is a long string uniquely associated with a particular app.
	AppID string `json:"application"`

	// CheckSum is the SHA512/224 hash of the executable, used to associate
	// tree data with a particular release.
	CheckSum string `json:"checksum"`

	// IP is the public IP address of the device on which we are running,
	// used to associate tree data with an estimated geographic location.
	IP string `json:"publicIP"`

	// UUID is a unique identifier for a particular tree.
	UUID string `json:"id"`

	// Time is the Unix epoch time (in milliseconds) at which a tree was
	// received.
	Time int64 `json:"timestamp"`

	// Tree represents the profile tree data generated by an agent.
	Tree  interface{} `json:"tree"`
	Error string      `json:"error,omitempty"`
}

// NewProfile creates a Profile for app out of raw message data.
func NewProfile(data []byte, app App) broker.Message {
	var p profile
	err := json.Unmarshal(data, &p)
	if err != nil {
		p.Error = err.Error()
	}
	p.IP = device.CurrentIP()
	p.UUID = uuid.NewV4().String()
	p.Time = time.Now().UnixNano() / 1000000 // milliseconds
	p.CheckSum = app.CheckSum()
	p.AppID = app.ID()
	return marshal(p, broker.Profile)
}
