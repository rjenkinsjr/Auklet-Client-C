package broker

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/eclipse/paho.mqtt.golang"

	backend "github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// MQTTProducer wraps an MQTT Client.
type MQTTProducer struct {
	c       Client
	org, id string
}

type token interface {
	Wait() bool
	Error() error
}

// wait turns Paho's async API into a sync API.
var wait = func(t token) error {
	t.Wait()
	return t.Error()
}

type Client interface {
	Connect() mqtt.Token
	Publish(string, byte, bool, interface{}) mqtt.Token
	Disconnect(uint)
}

// Config provides parameters for an MQTTProducer.
type Config struct {
	Creds  *backend.Credentials
	Client Client
}

// API consists of the backend interface needed to generate a Config.
type API interface {
	backend.Credentialer
	BrokerAddress() (string, error)
	Certificates() (*tls.Config, error)
}

// NewConfig returns a Config from the given API.
func NewConfig(api API) (Config, error) {
	creds, err := api.Credentials()
	if err != nil {
		return Config{}, err
	}

	addr, err := api.BrokerAddress()
	if err != nil {
		return Config{}, err
	}

	certs, err := api.Certificates()
	if err != nil {
		return Config{}, err
	}

	opt := mqtt.NewClientOptions()
	opt.AddBroker(addr)
	opt.SetTLSConfig(certs)
	opt.SetClientID(creds.ClientID)
	opt.SetCredentialsProvider(func() (string, string) {
		return creds.Username, creds.Password
	})

	return Config{
		Creds:  creds,
		Client: mqtt.NewClient(opt),
	}, nil
}

// NewMQTTProducer returns a new producer for the given input.
func NewMQTTProducer(cfg Config) (*MQTTProducer, error) {
	c := cfg.Client
	if err := wait(c.Connect()); err != nil {
		return nil, err
	}
	log.Print("producer: connected")
	return &MQTTProducer{
		c:   c,
		org: cfg.Creds.Org,
		id:  cfg.Creds.Username,
	}, nil
}

// Serve launches p, enabling it to send and receive messages.
func (p MQTTProducer) Serve(in MessageSource) {
	defer func() {
		p.c.Disconnect(0)
		log.Print("producer: disconnected")
	}()

	topic := map[Topic]string{
		Profile: "profiler",
		Event:   "events",
		Log:     "logs",
	}
	for k, v := range topic {
		topic[k] = fmt.Sprintf("c/%v/%v/%v", v, p.org, p.id)
	}

	for msg := range in.Output() {
		if err := wait(p.c.Publish(topic[msg.Topic], 1, false, []byte(msg.Bytes))); err != nil {
			errorlog.Print("producer:", err)
			continue
		}
		log.Printf("producer: sent %+q", msg.Bytes)
		msg.Remove()
	}
}
