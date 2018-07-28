package zapGL

import (
	"crypto/tls"
	"fmt"

	"net"

	"github.com/buger/jsonparser"
	"github.com/json-iterator/go"
	"go.uber.org/zap/zapcore"
)

// Transport represents a transport type enum
type Transport string

// Transport types
const (
	UDP Transport = "udp"
	TCP Transport = "tcp"
)

// Endpoint represents a graylog endpoint
type Endpoint struct {
	Transport Transport
	Address   string
	Port      uint
}

// Graylog represents an established graylog connection
type Graylog struct {
	Client    *net.Conn
	TLSClient *tls.Conn
}

// Message represents a GELF formated message
type Message struct {
	Version      string `json:"version"`
	Host         string `json:"host"`
	ShortMessage string `json:"short_message"`
	FullMessage  string `json:"full_message,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
	Level        uint   `json:"level,omitempty"`
	// Extra        map[string]string `json:"-"`
}

// map zapcore's log levels to standard syslog levels used by gelf, approximately.
var zapToSyslog = map[zapcore.Level]uint{
	zapcore.DebugLevel:  7,
	zapcore.InfoLevel:   6,
	zapcore.WarnLevel:   4,
	zapcore.ErrorLevel:  3,
	zapcore.DPanicLevel: 2,
	zapcore.PanicLevel:  2,
	zapcore.FatalLevel:  1,
}

// prepareMessage marshal the given message, add extra fields and append EOL symbols
func prepareMessage(m Message, extras map[string]string) ([]byte, error) {
	// Marshal the GELF message in order to get base JSON
	jsonMessage, err := jsoniter.Marshal(m)
	if err != nil {
		return []byte{}, err
	}

	// Parse JSON in order to dynamically edit it
	// c, err := gabs.ParseJSON(jsonMessage)

	if err != nil {
		return []byte{}, err
	}

	// Loop on extra fields and inject them into JSON
	for key, value := range extras {

		jsonMessage, err = jsonparser.Set(jsonMessage, jsonparser.StringToBytes("\""+value+"\""), fmt.Sprintf("_%s", key))
		if err != nil {
			return []byte{}, err
		}
	}

	// Append the \n\0 sequence to the given message in order to indicate
	// to graylog the end of the message

	finalMsg := map[string]interface{}{}

	err = jsoniter.Unmarshal(jsonMessage, &finalMsg)

	data := append(jsonMessage, '\n', byte(0))

	return data, nil
}
