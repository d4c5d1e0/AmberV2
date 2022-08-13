package logger

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	debug  *zap.Logger
)

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05.00000"))
}
func init() {
	cfg := zap.Config{
		Encoding:    "console",
		OutputPaths: []string{"stderr", "./amber.log"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			TimeKey:     "time",
			EncodeTime:  SyslogTimeEncoder,
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
	}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	logger, _ = cfg.Build()
	debugLogger()
}
func debugLogger() {
	cfg := zap.Config{
		Encoding:    "json",
		OutputPaths: []string{"./amber_debug.log"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			TimeKey:     "time",
			EncodeTime:  SyslogTimeEncoder,
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
	}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	debug, _ = cfg.Build()
}

func Debug(contents ...interface{}) {
	content := strings.TrimRight(fmt.Sprintln(contents...), " \n\r")
	debug.Debug(content)
}

func Info(contents ...interface{}) {
	content := strings.TrimRight(fmt.Sprintln(contents...), " \n\r")
	logger.Info(content)
}

func Error(contents ...interface{}) {
	content := strings.TrimRight(fmt.Sprintln(contents...), " \n\r")
	logger.Error(content)
	debug.Debug(content)
}
