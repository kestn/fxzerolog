package fxzerolog

import (
	"strings"

	"github.com/rs/zerolog"
	"go.uber.org/fx/fxevent"
)

var _ fxevent.Logger = (*ZerologLogger)(nil)

// ZerologLogger an Fx event logger that logs events by zerolog.
type ZerologLogger struct {
	Logger zerolog.Logger

	logLevel   zerolog.Level
	errorLevel *zerolog.Level
}

// UseLogLevel sets the level of non-error logs emitted by Fx to level.
func (l *ZerologLogger) UseLogLevel(level zerolog.Level) {
	l.logLevel = level
}

// UseErrorLevel sets the level of error logs emitted by Fx to level.
func (l *ZerologLogger) UseErrorLevel(level zerolog.Level) {
	l.errorLevel = &level
}

func (l *ZerologLogger) logEvent() *zerolog.Event {
	return l.Logger.WithLevel(l.logLevel)
}

func (l *ZerologLogger) errorLogEvent() *zerolog.Event {
	if l.errorLevel != nil {
		return l.Logger.WithLevel(*l.errorLevel)
	}

	return l.Logger.WithLevel(zerolog.ErrorLevel)
}

// LogEvent logs the given event to the provided Zerolog logger.
func (l *ZerologLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logEvent().
			Str("callee", e.FunctionName).
			Str("caller", e.CallerName).
			Msg("OnStart hook executing")
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.errorLogEvent().
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Err(e.Err).
				Msg("OnStart hook failed")
		} else {
			l.logEvent().
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Str("runtime", e.Runtime.String()).
				Msg("OnStart hook executed")
		}
	case *fxevent.OnStopExecuting:
		l.logEvent().
			Str("callee", e.FunctionName).
			Str("caller", e.CallerName).
			Msg("OnStop hook executing")
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.errorLogEvent().
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Err(e.Err).
				Msg("OnStop hook failed")
		} else {
			l.logEvent().
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Str("runtime", e.Runtime.String()).
				Msg("OnStop hook executed")
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			zEvent := l.errorLogEvent().
				Str("type", e.TypeName).
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace)
			maybeStringField(zEvent, "module", e.ModuleName).
				Err(e.Err).
				Msg("error encountered while applying options")
		} else {
			zEvent := l.logEvent().
				Str("type", e.TypeName).
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace)
			maybeStringField(zEvent, "module", e.ModuleName).
				Err(e.Err).
				Msg("supplied")
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			zEvent := l.logEvent().
				Str("constructor", e.ConstructorName).
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace)
			maybeStringField(zEvent, "module", e.ModuleName).
				Str("type", rtype)
			maybeBoolField(zEvent, "private", e.Private).
				Msg("provided")
		}
		if e.Err != nil {
			l.errorLogEvent().
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace).
				Err(e.Err).
				Msg("error encountered while applying options")
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			zEvent := l.logEvent().
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace)
			maybeStringField(zEvent, "module", e.ModuleName).
				Str("type", rtype).
				Msg("replaced")
		}
		if e.Err != nil {
			zEvent := l.errorLogEvent().
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace)
			maybeStringField(zEvent, "module", e.ModuleName).
				Err(e.Err).
				Msg("error encountered while replacing")
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			zEvent := l.logEvent().
				Str("decorator", e.DecoratorName).
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace)
			maybeStringField(zEvent, "module", e.ModuleName).
				Str("type", rtype).
				Msg("decorated")
		}
		if e.Err != nil {
			zEvent := l.errorLogEvent().
				Strs("stacktrace", e.StackTrace).
				Strs("moduletrace", e.ModuleTrace)
			maybeStringField(zEvent, "module", e.ModuleName).
				Err(e.Err).
				Msg("error encountered while applying options")
		}
	case *fxevent.Run:
		if e.Err != nil {
			zEvent := l.errorLogEvent().
				Str("name", e.Name).
				Str("kind", e.Kind)
			maybeStringField(zEvent, "module", e.ModuleName).
				Err(e.Err).
				Msg("error returned")
		} else {
			zEevent := l.logEvent().
				Str("name", e.Name).
				Str("kind", e.Kind).
				Str("runtime", e.Runtime.String())
			maybeStringField(zEevent, "module", e.ModuleName).
				Msg("run")
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		zEvent := l.logEvent().
			Str("function", e.FunctionName)
		maybeStringField(zEvent, "module", e.ModuleName).
			Msg("invoking")
	case *fxevent.Invoked:
		if e.Err != nil {
			zEvent := l.errorLogEvent().
				Err(e.Err).
				Str("stack", e.Trace).
				Str("function", e.FunctionName)
			maybeStringField(zEvent, "module", e.ModuleName).
				Msg("invoke failed")
		}
	case *fxevent.Stopping:
		l.logEvent().
			Str("signal", strings.ToUpper(e.Signal.String())).
			Msg("received signal")
	case *fxevent.Stopped:
		if e.Err != nil {
			l.errorLogEvent().
				Err(e.Err).
				Msg("stop failed")
		}
	case *fxevent.RollingBack:
		l.errorLogEvent().
			Err(e.StartErr).
			Msg("start failed, rolling back")
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.errorLogEvent().
				Err(e.Err).
				Msg("rollback failed")
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.errorLogEvent().
				Err(e.Err).
				Msg("start failed")
		} else {
			l.logEvent().
				Msg("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.errorLogEvent().
				Err(e.Err).
				Msg("custom logger initialization failed")
		} else {
			l.logEvent().
				Str("function", e.ConstructorName).
				Msg("initialized custom fxevent.Logger")
		}
	}
}

func maybeStringField(event *zerolog.Event, k, v string) *zerolog.Event {
	if len(v) == 0 {
		return event
	}

	return event.Str(k, v)
}

func maybeBoolField(event *zerolog.Event, k string, v bool) *zerolog.Event {
	if v {
		return event.Bool(k, v)
	}

	return event
}
