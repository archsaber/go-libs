package zapGL

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// GrayConnType Coonection type to graylog server, UDP or TCP
type GrayConnType string

// Connection types to use
const (
	GrayConnTCP = "TCP"
	GrayConnUDP = "UDP"
)

// Conf Config to setup the zap Core returned
type Conf struct {
	AppName             string        // app version to encode
	AppVersion          string        // app name to encode
	PanicOnFail         bool          // panic if unable to init the zapcore
	ZapMinLogLevel      zapcore.Level // min log level to send
	GrayLogHostname     string        // host address
	GrayLogConnPort     int           // connection port
	GrayLogConnType     GrayConnType  // udp or tcp
	GrayLogConnPoolInit int           // starting pool size
	GrayLogConnPoolMax  int           // max  Connection pool size
	GrayLogConnTimeOut  time.Duration // timeout in ms, default 2 seconds
	GrayLogBufferCap    int           // capacity of log message buffer
	GrayLogFlushTime    int           // flush timeout for log messages
}

func fillConfDefaults(conf Conf) Conf {

	if conf.AppName == "" {
		conf.AppName = "Sample-App"
	}
	if conf.AppVersion == "" {
		conf.AppVersion = "0.0.1"
	}

	if conf.GrayLogConnType == "" {
		conf.GrayLogConnType = GrayConnUDP
	}

	if conf.GrayLogHostname == "" {
		conf.GrayLogHostname = "localhost"
	}

	if conf.GrayLogConnPort == 0 {
		conf.GrayLogConnPort = 12201
	}

	if conf.ZapMinLogLevel == 0 {
		conf.ZapMinLogLevel = zapcore.InfoLevel
	}

	if conf.GrayLogConnPoolInit == 0 || conf.GrayLogConnPoolInit > 100 {
		conf.GrayLogConnPoolInit = 1
	}

	if conf.GrayLogConnPoolMax == 0 || conf.GrayLogConnPoolMax > 400 {
		conf.GrayLogConnPoolMax = 2
	}

	if conf.GrayLogConnTimeOut < 10*time.Millisecond || conf.GrayLogConnTimeOut > 16*time.Second {
		conf.GrayLogConnTimeOut = 2000 * time.Millisecond
	}

	if conf.GrayLogFlushTime < 50 {
		//  by default flush every 500 ms
		conf.GrayLogFlushTime = 500
	}

	if conf.GrayLogBufferCap < 0 {
		conf.GrayLogBufferCap = 1024
	}

	return conf
}
