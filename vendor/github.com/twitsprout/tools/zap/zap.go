package zap

import (
	"context"
	"io"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CtxValueFunc is a function that produces log keys from a context.
type CtxValueFunc func(ctx context.Context) (key string, val interface{}, ok bool)

// Level is the logging priority that messages will get logged at.
type Level int8

const (
	// DebugLevel is the logging priority for debug messaging, usually disabled in production.
	DebugLevel Level = 1
	// InfoLevel is the default logging priority.
	InfoLevel Level = 2
	// WarnLevel is the logging priority for important but not critical messaging.
	WarnLevel Level = 3
	// ErrorLevel is the logging priority for high-priority messaging.
	ErrorLevel Level = 4
)

// Config contains the configurable options for the Zap implementation.
type Config struct {
	App           string
	Version       string
	Out           io.Writer
	LogLevel      Level
	CallerKey     string
	LevelKey      string
	MessageKey    string
	TimeKey       string
	CtxValueFuncs []CtxValueFunc
}

// Zap implements the tools Logger interface using Uber's zap library.
type Zap struct {
	config Config
	logLvl zap.AtomicLevel
	logger *zap.SugaredLogger
}

var defaultConfig = Config{
	LogLevel:   InfoLevel,
	CallerKey:  "caller",
	LevelKey:   "level",
	MessageKey: "msg",
	TimeKey:    "time",
}

var levelMap = map[Level]zapcore.Level{
	DebugLevel: zapcore.DebugLevel,
	InfoLevel:  zapcore.InfoLevel,
	WarnLevel:  zapcore.WarnLevel,
	ErrorLevel: zapcore.ErrorLevel,
}

// NewFromConfig returns a new Zap logger using the provided configuration.
func NewFromConfig(config Config) *Zap {
	if config.LogLevel == 0 {
		config.LogLevel = defaultConfig.LogLevel
	}
	if config.CallerKey == "" {
		config.CallerKey = defaultConfig.CallerKey
	}
	if config.LevelKey == "" {
		config.LevelKey = defaultConfig.LevelKey
	}
	if config.MessageKey == "" {
		config.MessageKey = defaultConfig.MessageKey
	}
	if config.TimeKey == "" {
		config.TimeKey = defaultConfig.TimeKey
	}

	level, ok := levelMap[config.LogLevel]
	if !ok {
		level = zapcore.InfoLevel
	}
	logLvl := zap.NewAtomicLevel()
	logLvl.SetLevel(level)

	encConfig := zapcore.EncoderConfig{
		CallerKey:      config.CallerKey,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		LevelKey:       config.LevelKey,
		MessageKey:     config.MessageKey,
		TimeKey:        config.TimeKey,
	}

	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encConfig),
			zapcore.AddSync(config.Out),
			logLvl,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(
			zapcore.Field{
				Key:    "app",
				Type:   zapcore.StringType,
				String: config.App,
			},
			zapcore.Field{
				Key:    "v",
				Type:   zapcore.StringType,
				String: config.Version,
			},
		),
	)

	return &Zap{
		config: config,
		logLvl: logLvl,
		logger: logger.Sugar(),
	}
}

// New returns a new Zap logger using the provided version string.
func New(app, version string, out io.Writer) *Zap {
	return NewFromConfig(Config{
		App:     app,
		Version: version,
		Out:     out,
	})
}

// Debug logs a debug message.
func (z *Zap) Debug(msg string, keyVals ...interface{}) {
	z.logger.Debugw(msg, keyVals...)
}

// DebugCtx logs a contextual debug message.
func (z *Zap) DebugCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	keyVals = z.withContextVals(ctx, keyVals)
	z.Debug(msg, keyVals...)
}

// Error logs an error message.
func (z *Zap) Error(msg string, keyVals ...interface{}) {
	z.logger.Errorw(msg, keyVals...)
}

// ErrorCtx logs a contextual error message.
func (z *Zap) ErrorCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	keyVals = z.withContextVals(ctx, keyVals)
	z.Error(msg, keyVals...)
}

// Handler returns an HTTP handler to update the logging level.
func (z *Zap) Handler() http.Handler {
	return z.logLvl
}

// Info logs an info message.
func (z *Zap) Info(msg string, keyVals ...interface{}) {
	z.logger.Infow(msg, keyVals...)
}

// InfoCtx logs a contextual info message.
func (z *Zap) InfoCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	keyVals = z.withContextVals(ctx, keyVals)
	z.Info(msg, keyVals...)
}

// Level returns the current logging level string.
func (z *Zap) Level() string {
	return z.logLvl.Level().String()
}

// SetLevel sets the logging level to the provided string, returning an error
// if the level is invalid.
func (z *Zap) SetLevel(level string) error {
	return z.logLvl.UnmarshalText([]byte(level))
}

// Warn logs a warn message.
func (z *Zap) Warn(msg string, keyVals ...interface{}) {
	z.logger.Warnw(msg, keyVals...)
}

// WarnCtx logs a contextual warn message.
func (z *Zap) WarnCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	keyVals = z.withContextVals(ctx, keyVals)
	z.Warn(msg, keyVals...)
}

func (z *Zap) withContextVals(ctx context.Context, keyVals []interface{}) []interface{} {
	for _, f := range z.config.CtxValueFuncs {
		key, val, ok := f(ctx)
		if ok {
			keyVals = append(keyVals, key, val)
		}
	}

	return keyVals
}
