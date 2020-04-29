# mqttmux

This is a Go mux-like interface for handling mqtt messages, which uses the Paho
MQTT driver underneath. The aim for mqttmux is to provide a topic subscription
abstraction that resembles the Go HTTP Handler programming model.

Mqttmux is not a real multiplexer in the sense that it doesn't actually route
the messages during runtime. It simply registers topics and handlers in a Route,
and when `mqttmux.Init()` is called, all topic subscriptions are set (using Paho),
with the associated `HandlerFunc` as message callback function.

## Usage

Simply obtain a new mux instance using `New()`, passing in a reference to a
`mqtt.Client`. After that new handlers can be registered to topics using the
`Handle` method. The `mux.Init()` will subsequently do all topic subscriptions
with the Paho MQTT driver.

### Example

```Go
// Handle OS signals
interrupt := make(chan os.Signal, 1)
signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

// Get a mqtt muxer
mux := mqttmux.New(&mqttcli)

// Set a topic handler
mux.Handle("devices/:device_id/cmd", 1, func(m mqtt.Message, p mqttmux.MQTTParams) {
    deviceID, _ := p.Get("device_id")
    fmt.Printf("Received command: %s, for device: %s\n", string(m.Payload()), deviceID)
})

// Init() is non-blocking...
mux.Init()

// ...so we wait..
<-interrupt
```
