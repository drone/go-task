package logger

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	logger                = logrus.StandardLogger()
	closableHooks         []ClosableHook
	contextUpdatableHooks []ContextUpdatableHook
)

// ClosableHook is an interface for hooks that require a Close operation.
type ClosableHook interface {
	logrus.Hook
	Close() error
}

type ContextUpdatableHook interface {
	logrus.Hook
	UpdateContext(context map[string]string)
}

func GetLogger() *logrus.Logger {
	return logger
}

func getLogger() *logrus.Logger {
	return logger
}

// Configuration Methods

// SetReportCaller sets the log formatter.
func SetReportCaller(reportCaller bool) {
	getLogger().SetReportCaller(reportCaller)
}

// SetFormatter sets the log formatter.
func SetFormatter(formatter logrus.Formatter) {
	getLogger().SetFormatter(formatter)
}

// SetOutput sets the output destination for the logs.
func SetOutput(output io.Writer) {
	getLogger().SetOutput(output)
}

// AddHook adds a hook to the logger and tracks it if it's a ClosableHook.
func AddHook(hook logrus.Hook) {
	getLogger().AddHook(hook)
	if ch, ok := hook.(ClosableHook); ok {
		closableHooks = append(closableHooks, ch)
	}
	if ch, ok := hook.(ContextUpdatableHook); ok {
		contextUpdatableHooks = append(contextUpdatableHooks, ch)
	}
}

// CloseHooks closes all closable hooks.
func CloseHooks() error {
	var errorList []string
	for _, hook := range closableHooks {
		if err := hook.Close(); err != nil {
			errorList = append(errorList, err.Error())
		}
	}

	// If errorList is not empty, join the errors into a single error message
	if len(errorList) > 0 {
		return errors.New(strings.Join(errorList, "; "))
	}

	return nil
}

// UpdateContextInHooks updates all updatable hooks with extra fields.
func UpdateContextInHooks(context map[string]string) {
	for _, hook := range contextUpdatableHooks {
		hook.UpdateContext(context)
	}
}

// Log Methods

func WithError(ctx context.Context, err error) *logrus.Entry {
	return getLogger().WithContext(ctx).WithError(err)
}
func WithContext(ctx context.Context) *logrus.Entry {
	return getLogger().WithContext(ctx).WithContext(ctx)
}
func WithField(ctx context.Context, key string, value interface{}) *logrus.Entry {
	return getLogger().WithContext(ctx).WithField(key, value)
}
func WithFields(ctx context.Context, fields map[string]interface{}) *logrus.Entry {
	return getLogger().WithContext(ctx).WithFields(fields)
}
func WithTime(ctx context.Context, t time.Time) *logrus.Entry {
	return getLogger().WithContext(ctx).WithTime(t)
}

func Trace(ctx context.Context, args ...interface{})   { getLogger().WithContext(ctx).Trace(args...) }
func Debug(ctx context.Context, args ...interface{})   { getLogger().WithContext(ctx).Debug(args...) }
func Print(ctx context.Context, args ...interface{})   { getLogger().WithContext(ctx).Print(args...) }
func Info(ctx context.Context, args ...interface{})    { getLogger().WithContext(ctx).Info(args...) }
func Warn(ctx context.Context, args ...interface{})    { getLogger().WithContext(ctx).Warn(args...) }
func Warning(ctx context.Context, args ...interface{}) { getLogger().WithContext(ctx).Warning(args...) }
func Error(ctx context.Context, args ...interface{})   { getLogger().WithContext(ctx).Error(args...) }
func Panic(ctx context.Context, args ...interface{})   { getLogger().WithContext(ctx).Panic(args...) }
func Fatal(ctx context.Context, args ...interface{})   { getLogger().WithContext(ctx).Fatal(args...) }

func Tracef(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Tracef(format, args...)
}
func Debugf(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Debugf(format, args...)
}
func Printf(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Printf(format, args...)
}
func Infof(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Infof(format, args...)
}
func Warnf(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Warnf(format, args...)
}
func Warningf(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Warningf(format, args...)
}
func Errorf(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Errorf(format, args...)
}
func Panicf(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Panicf(format, args...)
}
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	getLogger().WithContext(ctx).Fatalf(format, args...)
}

func Traceln(ctx context.Context, args ...interface{}) { getLogger().WithContext(ctx).Traceln(args...) }
func Debugln(ctx context.Context, args ...interface{}) { getLogger().WithContext(ctx).Debugln(args...) }
func Println(ctx context.Context, args ...interface{}) { getLogger().WithContext(ctx).Println(args...) }
func Infoln(ctx context.Context, args ...interface{})  { getLogger().WithContext(ctx).Infoln(args...) }
func Warnln(ctx context.Context, args ...interface{})  { getLogger().WithContext(ctx).Warnln(args...) }
func Warningln(ctx context.Context, args ...interface{}) {
	getLogger().WithContext(ctx).Warningln(args...)
}
func Errorln(ctx context.Context, args ...interface{}) { getLogger().WithContext(ctx).Errorln(args...) }
func Panicln(ctx context.Context, args ...interface{}) { getLogger().WithContext(ctx).Panicln(args...) }
func Fatalln(ctx context.Context, args ...interface{}) { getLogger().WithContext(ctx).Fatalln(args...) }
