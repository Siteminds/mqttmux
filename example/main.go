package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Siteminds/mqttmux"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Turn on debug logging for this example
	log.SetLevel(log.DebugLevel)

	// Handle OS signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Get MQTT Connection
	uri, err := url.Parse("tcp://localhost:1883")
	if err != nil {
		log.Fatal(err)
	}
	mqttcli := connect("mqttmux-example", uri)
	log.WithField("uri", uri).Info("connected to MQTT broker")

	// Get a mqtt muxer
	mux := mqttmux.New(&mqttcli)

	// Set a topic handler
	mux.Handle("devices/:device_id/cmd", 1, deviceCMDHandler)

	// Init() is non-blocking...
	mux.Init()

	// ...so we wait until we get a stop signal
	select {
	case s := <-interrupt:
		log.WithFields(log.Fields{"signal": s}).Info("received OS signal")
		mqttcli.Disconnect(2000)
	}
	log.Info("Done.")
}

// This is our HandlerFunc
func deviceCMDHandler(m mqtt.Message, p mqttmux.MQTTParams) {
	deviceID, _ := p.Get("device_id")
	log.Infof("Received command: %s, for device: %s\n", string(m.Payload()), deviceID)
}

// Convenience function for setting up MQTT connection
func connect(clientID string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientID, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatalf("error connecting to MQTT broker: %v", err)
	}
	return client
}

// Convenience function for setting the MQTT client options
func createClientOptions(clientID string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s", uri.Scheme, uri.Host))
	opts.SetUsername(uri.User.Username())
	if password, isSet := uri.User.Password(); isSet {
		opts.SetPassword(password)
	}
	opts.SetClientID(clientID)
	opts.SetProtocolVersion(4)
	return opts
}
