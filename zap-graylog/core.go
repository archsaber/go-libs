package zapGL

import (
	"context"
	"errors"
	"fmt"

	// "log"
	"os"
	"strconv"
	"time"

	log "github.com/kr/pretty"

	"github.com/Broadroad/gpool"
	"github.com/json-iterator/go"
	"go.uber.org/atomic"
	"go.uber.org/zap/zapcore"
)

// Core returned for use with zap
type Core struct {
	zapcore.LevelEnabler                 // filter by severity
	conf                 *Conf           // gray-zap config
	connPool             *gpool.Pool     // connection pool
	Context              []zapcore.Field // zap context fields
	encoder              zapcore.Encoder // zap field encoder
	msgBuffer            chan []byte     // messages buffer
	lastFlush            *atomic.Int64   // Timestamp of last flush
	flushTimer           int64           // max time to a keep a msg in buffer
	flushLock            *atomic.Bool    // flush lock
}

// MaxPacketSize max allowed packet size in bytes
const MaxPacketSize = 2 * 1024 * 1024

// New create a from the give conf
func New(conf Conf) *Core {

	conf = fillConfDefaults(conf)

	encoderConfigs := zapcore.EncoderConfig{
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfigs)
	lastFlush := atomic.NewInt64(giveTSinMS())
	flushLock := atomic.NewBool(false)

	return &Core{
		LevelEnabler: conf.ZapMinLogLevel,
		conf:         &conf,
		connPool:     makeTCPConnPool(conf),
		encoder:      encoder,
		msgBuffer:    make(chan []byte, 4096),
		lastFlush:    lastFlush,
		flushLock:    flushLock,
		flushTimer:   int64(conf.GrayLogFlushTime),
		Context:      []zapcore.Field{},
	}
}

// With adds structured context to the logger.
func (core *Core) With(fields []zapcore.Field) zapcore.Core {

	core.Context = append(core.Context, fields...)
	// Done.
	return core
}

// Check determines whether the supplied entry should be logged.
func (core *Core) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {

	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}

	return checked
}

// Write writes messages to the configured Graylog endpoint.
func (core *Core) Write(entry zapcore.Entry, fields []zapcore.Field) error {

	if core.connPool == nil {
		return errors.New("Unintialized core")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	extraFields := map[string]string{
		"logger_name": entry.LoggerName,
		"app_name":    core.conf.AppName,
		"app_version": core.conf.AppVersion,
	}

	if entry.Caller.Defined {
		extraFields["file"] = entry.Caller.File
		extraFields["line"] = strconv.Itoa(entry.Caller.Line)
		extraFields["package"] = entry.Caller.TrimmedPath()
	}

	// the order here is important,
	// as fields supplied at the log site should overwrite fields supplied in the context
	for _, field := range core.Context {
		extraFields[field.Key] = field.String
	}

	// Encode the zap fields from fields to JSON with proper types.
	buf, err := core.encoder.EncodeEntry(entry, fields)

	if err != nil {
		if core.conf.PanicOnFail {
			panic(err)
		} else {
			log.Println("Unable to decode zap fields", err.Error())
		}
	}

	// Unmarshal the JSON into a map.
	m := make(map[string]interface{})
	if err = jsoniter.Unmarshal(buf.Bytes(), &m); err != nil {
		return err
	}

	// Parse the map and return only strings.
	for k, v := range m {
		extraFields[k] = fmt.Sprintf("%v", v)
	}

	ts := entry.Time.Unix()
	msg := Message{
		Version:      "1.1",
		Host:         hostname,
		ShortMessage: entry.Message,
		FullMessage:  entry.Stack,
		Timestamp:    ts,
		Level:        zapToSyslog[entry.Level],
		// Extra:        extraFields,
	}

	msgBytes, err := prepareMessage(msg, extraFields)

	// core.sendPacket(msgBytes)

	if err != nil {
		log.Println(err.Error())
	}

	core.pushToBuffer(msgBytes)

	return err
}

// Sync is a no-op.
func (core *Core) Sync() error {

	return nil
}

func (core *Core) pushToBuffer(b []byte) {
	select {
	case core.msgBuffer <- b:
	default:
		log.Println("gray-zap message buffer full, entry dropped from sending")
	}

	// by this time msg is guaranteed in buffer or dropped
	if core.flushLock.CAS(false, true) {
		go core.flush()
	}
}

func (core *Core) flush() bool {
	var err error
	// if enough messages then flush anyway
	if len(core.msgBuffer) > 1000 {
		_, err = core.sendBuffer()
	} else {
		now := giveTSinMS()

		lastFlush := core.lastFlush.Load()

		timeToFlush := core.flushTimer - (now - lastFlush)

		if timeToFlush > 0 && timeToFlush < core.flushTimer {
			time.Sleep(time.Duration(timeToFlush) * time.Millisecond)
		}

		_, err = core.sendBuffer()
	}

	if err != nil {
		log.Println("Error sending to graylog-server : ", err.Error())
	}

	core.lastFlush.Store(giveTSinMS())

	return core.flushLock.CAS(true, false)
}

func (core *Core) sendBuffer() (int, error) {
	packet := []byte{}

	for len(core.msgBuffer) > 0 {
		select {
		case b := <-core.msgBuffer:
			packet = append(packet, b...)
		default:
			continue
		}
		if len(packet) > MaxPacketSize {
			core.sendPacket(packet)
			packet = []byte{}
		}
	}

	if len(packet) > 0 {
		return core.sendPacket(packet)
	}

	return 0, nil
}

func (core *Core) sendPacket(packet []byte) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), core.conf.GrayLogConnTimeOut) //3second timeout
	defer cancel()
	conn, err := (*core.connPool).BlockingGet(ctx)

	if err != nil {
		log.Println("Get error:", err)
	}

	defer conn.Close()
	return conn.Write(packet)
}

func giveTSinMS() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
