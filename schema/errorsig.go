// +build linux

package schema

import (
	"encoding/json"
	"syscall"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/device"
)

// ErrorSig represents the exit of an app in which libauklet handled an "error
// signal" and produced a stacktrace.
type ErrorSig struct {
	AppID string `json:"application"`
	// CheckSum is the SHA512/224 hash of the executable, used to associate
	// event data with a particular release.
	CheckSum string `json:"checksum"`

	// IP is the public IP address of the device on which we are running,
	// used to associate event data with an estimated geographic location.
	IP string `json:"publicIP"`

	// UUID is a unique identifier for a particular event.
	UUID string `json:"id"`

	// Time is the time at which the event was received.
	Time time.Time `json:"timestamp"`

	// Status is the exit status of the application as accessible through
	// App.Wait.
	Status int `json:"exitStatus"`

	// Signal is an integer value provided by libauklet. In JSON output, it
	// is represented as a string.
	Signal sig `json:"signal"`

	// Trace is a stacktrace provided by libauklet.
	Trace      json.RawMessage `json:"stackTrace"`
	MacHash    string          `json:"macAddressHash"`
	Metrics    device.Metrics  `json:"systemMetrics"`
	kafkaTopic string
}

// NewErrorSig creates an ErrorSig for app out of JSON data. It assumes that
// app.Wait() has returned.
func NewErrorSig(data []byte, app *app.App, topic string) (e ErrorSig, err error) {
	err = json.Unmarshal(data, &e)
	if err != nil {
		return
	}
	e.AppID = app.AppID
	e.CheckSum = app.CheckSum
	e.IP = device.CurrentIP()
	e.UUID = uuid.NewV4().String()
	e.Time = time.Now()
	e.Status = app.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	e.MacHash = device.MacHash
	e.Metrics = device.GetMetrics()
	e.kafkaTopic = topic
	return
}

// Topic returns the Kafka topic to which p should be sent.
func (e ErrorSig) Topic() string {
	return e.kafkaTopic
}

// Bytes returns the ErrorSig as a byte slice.
func (e ErrorSig) Bytes() ([]byte, error) {
	return json.MarshalIndent(e, "", "\t")
}

type sig syscall.Signal

// String returns s represented as a human-readable string.
func (s sig) String() string {
	return syscall.Signal(s).String()
}

// MarshalText encodes sig as a human-readable string.
func (s sig) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}