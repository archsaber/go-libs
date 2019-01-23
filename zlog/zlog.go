package zlog

import (
	"fmt"
	"os"

	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/tchap/zapext/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GiveLogger impl
func GiveLogger() *zap.SugaredLogger {
	return zap.L().Sugar()
}

var (
	cutPathLen     int
	releaseVersion string
)

func init() {
	archEnv := os.Getenv("ARCH_ENV")
	if archEnv == "DEV" && os.Getenv("SENTRY") != "true" {
		logger, _ := zap.NewDevelopment()
		zap.ReplaceGlobals(logger)
		return
	}

	consoleCore := giveConsoleCore()
	cores := []zapcore.Core{consoleCore}

	if os.Getenv("SENTRY") == "true" {
		cores = append(cores, giveSentryCore())
	}

	teeCore := zapcore.NewTee(cores...)
	dLogger := zap.New(teeCore, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	zap.ReplaceGlobals(dLogger)
}

func giveConsoleCore() zapcore.Core {
	cutPathLen = len(os.Getenv("PWD")) + 5
	pathEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {

		filePath := caller.String()

		if len(filePath) > cutPathLen {
			filePath = filePath[cutPathLen:]
		}

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

	consoleConf := zap.NewDevelopmentEncoderConfig()
	consoleConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleConf.EncodeTime = timeEncoder
	consoleConf.EncodeCaller = pathEncoder

	pDir := os.Getenv("PROJECT_DIR")
	logDir := os.Getenv("HOME") + "/log/"
	if pDir != "" {
		logDir = pDir + "/log/"
	}

	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		panic(err)
	}

	tNow := time.Now()
	datetimePrefix := fmt.Sprintf("%d_%02d_%02d", tNow.Year(), tNow.Month(), tNow.Day())
	pathLog := logDir + os.Getenv("ARCH_ENV") + "_" + datetimePrefix + ".log"
	fLog, err := os.OpenFile(pathLog, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleConf),
		zapcore.Lock(zapcore.NewMultiWriteSyncer(fLog)),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.DebugLevel
		}),
	)
}

func giveSentryCore() zapcore.Core {
	sentryDSN := os.Getenv("SENTRY_DSN")
	logLevel := zapcore.ErrorLevel
	if os.Getenv("ARCH_ENV") == "DEV" {
		sentryDSN = os.Getenv("DEV_SENTRY_DSN")
		logLevel = zapcore.InfoLevel
	}

	logLevelEnv := os.Getenv("SENTRY_LOG_LEVEL")
	if logLevelEnv == "WARN" {
		logLevel = zapcore.WarnLevel
	}

	client, err := raven.New(sentryDSN)
	if err != nil {
		panic(err)
	}

	client.SetEnvironment(os.Getenv("ARCH_ENV"))
	client.SetRelease(releaseVersion)

	return zapsentry.NewCore(logLevel, client)
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
