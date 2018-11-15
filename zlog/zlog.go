package zlog

import (
	"fmt"
	"os"

	"time"

	"github.com/getsentry/raven-go"
	"github.com/tchap/zapext/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GiveLogger impl
func GiveLogger() *zap.SugaredLogger {
	return zap.L().Sugar()
}

var cutPathLen int

func init() {

	if os.Getenv("SENTRY") == "true" {
		initSentryZap()
		return
	}
}

func initSentryZap() {

	sentryDSN := os.Getenv("SENTRY_DSN")
	archENV := os.Getenv("ARCH_ENV")

	client, err := raven.New(sentryDSN)
	if err != nil {
		panic(err)
	}

	setnryCore := zapsentry.NewCore(zapcore.ErrorLevel, client)

	// -------------------------

	cutPathLen = len(os.Getenv("PWD")) + 5
	pathEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {

		filePath := caller.String()[cutPathLen:]

		var pathString string
		if len(filePath) > 32 {
			diff := len(filePath) - 32
			pathString = fmt.Sprintf("[ %-32s ]", filePath[diff:])
		} else {
			pathString = fmt.Sprintf("[ %-32s ]", filePath)
		}

		enc.AppendString(pathString)
	}

	tLoc, _ := time.LoadLocation("Asia/Kolkata")

	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(tLoc).Format("2006 Jan _2 15:04:05"))
	}

	// -------------------------

	consoleConf := zap.NewDevelopmentEncoderConfig()

	consoleConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleConf.EncodeTime = timeEncoder
	consoleConf.EncodeCaller = pathEncoder

	// debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	// 	return lvl == zapcore.DebugLevel
	// })

	allPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})

	home := os.Getenv("HOME")
	pDir := os.Getenv("PROJECT_DIR")

	logDir := home

	if pDir != "" {
		logDir = pDir
	}

	datetimePrefix := time.Now().Format("2006-01-28-15-04-05")

	pathLog := logDir + "/log/" + archENV + datetimePrefix + ".log"
	// pathDebug := home + "/log/aws-nom.debug"

	fLog, err := os.OpenFile(pathLog, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	// fDebug, err := os.OpenFile(pathDebug, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	if err != nil {
		panic(err)
	}

	consoleLogOut := zapcore.Lock(zapcore.NewMultiWriteSyncer(fLog))
	// consoleDebugOut := zapcore.Lock(zapcore.NewMultiWriteSyncer(fDebug))

	consoleEncoder := zapcore.NewConsoleEncoder(consoleConf)

	core := zapcore.NewTee(
		// zapcore.NewCore(consoleEncoder, consoleDebugOut, debugLevel),
		zapcore.NewCore(consoleEncoder, consoleLogOut, allPriority),
		setnryCore,
	)

	logger := zap.New(core)

	zap.ReplaceGlobals(logger)

	fmt.Println("Using zap with sentry")

}

// Cl impl
func Cl() *zap.SugaredLogger {
	return zap.L().Sugar()
}

// Info log info event
func Info(msg string, keysAndValues ...interface{}) {
	zap.L().Sugar().Infow(msg, keysAndValues...)
}

// Warn log warn event
func Warn(msg string, keysAndValues ...interface{}) {
	zap.L().Sugar().Warnw(msg, keysAndValues...)
}

// Error log error events
func Error(msg string, keysAndValues ...interface{}) {
	zap.L().Sugar().Errorw(msg, keysAndValues...)
}

// Debug log Debug events
func Debug(msg string, keysAndValues ...interface{}) {
	zap.L().Sugar().Debugw(msg, keysAndValues...)
}

// Fatal log Fatal events
func Fatal(msg string, keysAndValues ...interface{}) {
	zap.L().Sugar().Fatalw(msg, keysAndValues...)
}

// Panic log panic events
func Panic(msg string, keysAndValues ...interface{}) {
	zap.L().Sugar().Panicw(msg, keysAndValues...)
}
