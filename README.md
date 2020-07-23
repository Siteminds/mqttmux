[![Go Report Card](https://goreportcard.com/badge/github.com/Siteminds/mqttmux)](https://goreportcard.com/report/github.com/Siteminds/mqttmux)
[![Github Action](https://github.com/Siteminds/mqttmux/workflows/Go/badge.svg)](https://github.com/Siteminds/mqttmux/actions?query=workflow%3AGo)

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

Or checkout the `example` directory for a more extensive, working example:

```shell
> cd example
> go run main.go
```

Output:

```shell
INFO[0000] connected to MQTT broker                      uri="tcp://localhost:1883"
INFO[0000] mux setting topic subscriptions
DEBU[0000] setting subscription                          handler=main.deviceCMDHandler qos=1 topic=devices/+/cmd
INFO[0000] mux done
DEBU[0009] mux: extract parameter values
DEBU[0009] mux: execute handler
INFO[0009] Received command: HELLO, for device: 12345
^CINFO[0017] received OS signal                            signal=interrupt
INFO[0017] Done.
```
