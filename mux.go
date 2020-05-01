package mqttmux

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

var paramRegex = regexp.MustCompile(`(\:\w+)\/*`)

// HandlerFunc is the function (proto-)type for message
// handling routines
type HandlerFunc func(mqtt.Message, MQTTParams)

// Route represents a topic+handler combo
type Route struct {
	SubPattern string
	QOS        byte
	Params     map[int]string
	Handler    HandlerFunc
}

// Mux is our router/multiplexer
type Mux struct {
	mu     sync.RWMutex
	re     *regexp.Regexp
	cli    *mqtt.Client
	Routes map[string]*Route
}

// New returns a new initialized Mux instance
func New(cli *mqtt.Client) *Mux {
	return &Mux{
		re:     paramRegex,
		cli:    cli,
		Routes: make(map[string]*Route),
	}
}

// MQTTParams contains a map with key/values
type MQTTParams map[string]string

// Get returns the value for a specific key
// and a boolean indicating if the key was
// found
func (p MQTTParams) Get(key string) (string, bool) {
	val, ok := p[key]
	return val, ok
}

// Set puts a specific value on a key
func (p MQTTParams) Set(key, value string) {
	p[key] = value
}

// Handle will register a new handler for a given
// topic pattern to the Mux
func (m *Mux) Handle(pattern string, qos byte, handler HandlerFunc) {
	m.mu.Lock()
	m.Routes[pattern] = &Route{
		SubPattern: strings.TrimSuffix(m.re.ReplaceAllString(pattern, `+/`), "/"),
		Params:     extractParams(pattern),
		QOS:        qos,
		Handler:    handler,
	}
	m.mu.Unlock()
}

// Init will make the mux register all subscriptions
// message channel, and dispatch messages to the
// correct handler
func (m *Mux) Init() {
	log.Info("mux setting topic subscriptions")
	cli := *m.cli
	for _, r := range m.Routes {
		// Debug logging
		log.WithFields(log.Fields{
			"topic":   r.SubPattern,
			"qos":     r.QOS,
			"handler": handlerName(r.Handler),
		}).Debug("setting subscription")

		// Set subscription
		if token := cli.Subscribe(r.SubPattern, r.QOS, mqttMsgHandlerFunc(r)); token.Wait() && token.Error() != nil {
			log.Errorf("error subscribing to %s: %v", r.SubPattern, token.Error())
		}
	}
	log.Info("mux done")
}

// extractParams will extract all parameters from the topic pattern
func extractParams(pattern string) map[int]string {
	params := make(map[int]string)
	parts := strings.Split(pattern, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			// we found a parameter
			params[i] = strings.TrimPrefix(p, ":")
		}
	}
	return params
}

// extractParamValues will take a message, and extract any parameters
// from the topic pattern (if needed)
func extractParamValues(r *Route, topic string) MQTTParams {
	subTopics := strings.Split(topic, "/")
	result := make(map[string]string)
	for i, p := range r.Params {
		result[p] = subTopics[i]
	}
	return result
}

// mqttMsgHandlerFunc is a wrapper that returns a mqtt.MessageHandler
// function, as required by mqtt.Subscribe. It first extracts the parameter
// values from the message topic, and passes it along with the message
// to the actual handler that was configured.
func mqttMsgHandlerFunc(r *Route) func(mqtt.Client, mqtt.Message) {
	return func(cli mqtt.Client, msg mqtt.Message) {
		log.Debug("mux: extract parameter values")
		p := extractParamValues(r, msg.Topic())
		log.Debug("mux: execute handler")

		// async
		go r.Handler(msg, p)
	}
}

// handlerName uses reflection to get the name of the
// handler function (used for debug logging)
func handlerName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
