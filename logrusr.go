package log

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type logrAdapter struct {
	entry  *logrus.Entry
	logger *logrus.Logger
	Level  logrus.Level
}

var _ logr.Logger = &logrAdapter{}

func New(logger *logrus.Logger) (logr.Logger, error) {
	return &logrAdapter{
		entry:  logrus.NewEntry(logger),
		logger: logger,
		Level:  logrus.InfoLevel,
	}, nil
}

func (l *logrAdapter) Info(msg string, keysAndValues ...interface{}) {
	var entry *logrus.Entry
	if len(keysAndValues) > 0 {
		entry = l.entry.WithFields(parseKeyValuesToLogrusFields(keysAndValues))
	} else {
		entry = l.entry
	}

	switch l.Level {
	case logrus.TraceLevel:
		entry.Trace(msg)
	case logrus.DebugLevel:
		entry.Debug(msg)
	case logrus.WarnLevel:
		entry.Warn(msg)
	default:
		entry.Info(msg)
	}
}

func (l *logrAdapter) Enabled() bool {
	return l.logger.IsLevelEnabled(l.Level)
}

func (l *logrAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	var entry *logrus.Entry
	if len(keysAndValues) > 0 {
		entry = l.entry.WithFields(parseKeyValuesToLogrusFields(keysAndValues))
	} else {
		entry = l.entry
	}

	entry.WithError(err).Error(msg)
}

func (l *logrAdapter) V(level int) logr.InfoLogger {
	// level must not be negative
	if level < 0 {
		panic(errors.Errorf("log level must not be negative: %d", level))
	}

	// 0 : Info, 1... : Debug, Trace, ...
	newLevel := logrus.Level(level + 4)
	if newLevel.String() == "unknown" {
		newLevel = logrus.TraceLevel
	}

	if l.Level == newLevel {
		return l
	}

	copyEntry := *l.entry
	return &logrAdapter{
		entry: &copyEntry,
		Level: newLevel,
	}
}

func (l *logrAdapter) WithValues(keysAndValues ...interface{}) logr.Logger {
	if len(keysAndValues) <= 0 {
		return l
	}

	return &logrAdapter{
		entry: l.entry.WithFields(parseKeyValuesToLogrusFields(keysAndValues)),
		Level: l.Level,
	}
}

func (l *logrAdapter) WithName(name string) logr.Logger {
	if len(name) <= 0 {
		return l
	}

	return &logrAdapter{
		entry: l.entry.WithField("name", name),
		Level: l.Level,
	}
}

func parseKeyValuesToLogrusFields(keysAndValues []interface{}) logrus.Fields {
	fields := logrus.Fields{}
	var trailKey string
	for _, kv := range keysAndValues {
		if len(trailKey) > 0 {
			fields[trailKey] = kv
			trailKey = ""
		} else {
			trailKey = fmt.Sprintf("%v", kv)
		}
	}
	if len(trailKey) > 0 {
		fields[trailKey] = ""
	}

	return fields
}
