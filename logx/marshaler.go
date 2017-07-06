package logx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// these are words that are used by logstash so we have to change its name when
// they are processed.
var reservedWords = map[string]struct{}{
	"error": struct{}{},
	"type":  struct{}{},
}

// Marshaler defines the method to marshal entries
type Marshaler interface {
	Marshal(entry *entry) ([]byte, error)
}

// DummyMarshaler does nothing
type DummyMarshaler int

// Marshal returns the info encoded to be readable by humans
func (l DummyMarshaler) Marshal(entry *entry) ([]byte, error) {
	return nil, nil
}

// HumanMarshaler formats a log in a human-readable form
type HumanMarshaler int

// Marshal returns the info encoded to be readable by humans
func (l HumanMarshaler) Marshal(entry *entry) ([]byte, error) {
	var buffer bytes.Buffer
	if entry.time != nil {
		_, _ = buffer.WriteString(entry.time.Format("2006-01-02 15:04:05"))
		_, _ = buffer.WriteString(" ")
	}
	_, _ = buffer.WriteString(entry.level.String())
	_, _ = buffer.WriteString(" ")
	_, _ = buffer.WriteString(entry.message)
	if len(entry.fields) > 0 {
		_, _ = buffer.WriteString(" FIELDS")
		for _, field := range entry.fields {
			_, _ = buffer.WriteString(" ")
			_, _ = buffer.WriteString(field.Key)
			_, _ = buffer.WriteString("=")
			_, _ = buffer.WriteString(fmt.Sprintf("%v", field.Value))
		}
	}
	_, _ = buffer.WriteString("\n")
	return buffer.Bytes(), nil
}

// LogstashMarshaler marshalls the data to a logstash-compatible JSON
type LogstashMarshaler struct {
	channel     string
	product     string
	application string
	hostname    string
}

// NewLogstashMarshaler is the constructor of the concrete type.
func NewLogstashMarshaler(channel, product, application string) *LogstashMarshaler {
	return &LogstashMarshaler{
		channel:     channel,
		product:     product,
		application: application,
		hostname:    hostname,
	}
}

// Marshal returns the info encoded in the logstash format (JSON with special fields)
func (l *LogstashMarshaler) Marshal(entry *entry) ([]byte, error) {
	data := make(map[string]interface{})
	// logstash ones
	data["@version"] = 1
	if entry.time != nil {
		data["@timestamp"] = entry.time.Format(time.RFC3339)
	}
	data["severity"] = entry.level.String()
	data["message"] = entry.message
	// mandatory SP ones
	data["app_server_name"] = l.hostname
	data["channel"] = l.channel
	data["application"] = l.application
	data["product"] = l.product
	// rest
	for _, field := range entry.fields {
		value := fmt.Sprintf("%v", field.Value)
		if _, ok := reservedWords[field.Key]; ok {
			data[fmt.Sprintf("%sx", field.Key)] = value
		} else {
			data[field.Key] = value
		}
	}

	encodedData, err := json.Marshal(data)
	// as the std logger does
	if len(encodedData) == 0 || encodedData[len(encodedData)-1] != '\n' {
		encodedData = append(encodedData, '\n')
	}
	return encodedData, err
}
