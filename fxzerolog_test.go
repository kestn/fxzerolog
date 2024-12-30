package fxzerolog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxevent"
)

type zerologObservableEntry struct {
	record []byte
}

func (s zerologObservableEntry) all() map[string]any {
	fields := map[string]any{}

	if err := json.Unmarshal(s.record, &fields); err != nil {
		panic(err)
	}

	return fields
}

func (s zerologObservableEntry) Fields() map[string]any {
	fields := s.all()

	delete(fields, zerolog.MessageFieldName)
	delete(fields, zerolog.LevelFieldName)

	return fields
}

func (s zerologObservableEntry) Level() string {
	return s.all()[zerolog.LevelFieldName].(string)
}

func (s zerologObservableEntry) Message() string {
	return s.all()[zerolog.MessageFieldName].(string)
}

type zerologObservableLogger struct {
	bytes.Buffer
	entries []zerologObservableEntry
}

func (s *zerologObservableLogger) TakeAll() []zerologObservableEntry {
	b, err := io.ReadAll(&s.Buffer)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(b), string('\n'))
	s.entries = make([]zerologObservableEntry, 0, len(lines)-1)

	for _, line := range lines[:len(lines)-1] {
		s.entries = append(s.entries, zerologObservableEntry{record: []byte(line)})
	}

	return s.entries
}

func newZerologObservableLogger(level zerolog.Level) (zerolog.Logger, *zerologObservableLogger) {
	writer := &zerologObservableLogger{}
	logger := zerolog.New(writer)
	zerolog.SetGlobalLevel(level)
	return logger, writer
}

func TestZerologLogger(t *testing.T) {
	t.Parallel()

	someError := errors.New("some error")

	tests := []struct {
		name        string
		give        fxevent.Event
		wantMessage string
		wantFields  map[string]any
	}{
		{
			name: "OnStartExecuting",
			give: &fxevent.OnStartExecuting{
				FunctionName: "hook.onStart",
				CallerName:   "bytes.NewBuffer",
			},
			wantMessage: "OnStart hook executing",
			wantFields: map[string]any{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStart",
			},
		},
		{
			name: "OnStopExecuting",
			give: &fxevent.OnStopExecuting{
				FunctionName: "hook.onStop1",
				CallerName:   "bytes.NewBuffer",
			},
			wantMessage: "OnStop hook executing",
			wantFields: map[string]any{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStop1",
			},
		},
		{
			name: "OnStopExecuted/Error",
			give: &fxevent.OnStopExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Err:          fmt.Errorf("some error"),
			},
			wantMessage: "OnStop hook failed",
			wantFields: map[string]any{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStart1",
				"error":  "some error",
			},
		},
		{
			name: "OnStopExecuted",
			give: &fxevent.OnStopExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Runtime:      time.Millisecond * 3,
			},
			wantMessage: "OnStop hook executed",
			wantFields: map[string]any{
				"caller":  "bytes.NewBuffer",
				"callee":  "hook.onStart1",
				"runtime": "3ms",
			},
		},
		{
			name: "OnStartExecuted/Error",
			give: &fxevent.OnStartExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Err:          fmt.Errorf("some error"),
			},
			wantMessage: "OnStart hook failed",
			wantFields: map[string]any{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStart1",
				"error":  "some error",
			},
		},
		{
			name: "OnStartExecuted",
			give: &fxevent.OnStartExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Runtime:      time.Millisecond * 3,
			},
			wantMessage: "OnStart hook executed",
			wantFields: map[string]any{
				"caller":  "bytes.NewBuffer",
				"callee":  "hook.onStart1",
				"runtime": "3ms",
			},
		},
		{
			name: "Supplied",
			give: &fxevent.Supplied{
				TypeName:    "*bytes.Buffer",
				StackTrace:  []string{"main.main", "runtime.main"},
				ModuleTrace: []string{"main.main"},
			},
			wantMessage: "supplied",
			wantFields: map[string]any{
				"type":        "*bytes.Buffer",
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
			},
		},
		{
			name: "Supplied/Error",
			give: &fxevent.Supplied{
				TypeName:    "*bytes.Buffer",
				StackTrace:  []string{"main.main", "runtime.main"},
				ModuleTrace: []string{"main.main"},
				Err:         someError,
			},
			wantMessage: "error encountered while applying options",
			wantFields: map[string]any{
				"type":        "*bytes.Buffer",
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"error":       "some error",
			},
		},
		{
			name: "Provide",
			give: &fxevent.Provided{
				ConstructorName: "bytes.NewBuffer()",
				StackTrace:      []string{"main.main", "runtime.main"},
				ModuleTrace:     []string{"main.main"},
				ModuleName:      "myModule",
				OutputTypeNames: []string{"*bytes.Buffer"},
				Private:         false,
			},
			wantMessage: "provided",
			wantFields: map[string]any{
				"constructor": "bytes.NewBuffer()",
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"type":        "*bytes.Buffer",
				"module":      "myModule",
			},
		},
		{
			name: "PrivateProvide",
			give: &fxevent.Provided{
				ConstructorName: "bytes.NewBuffer()",
				StackTrace:      []string{"main.main", "runtime.main"},
				ModuleTrace:     []string{"main.main"},
				ModuleName:      "myModule",
				OutputTypeNames: []string{"*bytes.Buffer"},
				Private:         true,
			},
			wantMessage: "provided",
			wantFields: map[string]any{
				"constructor": "bytes.NewBuffer()",
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"type":        "*bytes.Buffer",
				"module":      "myModule",
				"private":     true,
			},
		},
		{
			name: "Provide/Error",
			give: &fxevent.Provided{
				StackTrace:  []string{"main.main", "runtime.main"},
				ModuleTrace: []string{"main.main"},
				Err:         someError,
			},
			wantMessage: "error encountered while applying options",
			wantFields: map[string]any{
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"error":       "some error",
			},
		},
		{
			name: "Replace",
			give: &fxevent.Replaced{
				ModuleName:      "myModule",
				StackTrace:      []string{"main.main", "runtime.main"},
				ModuleTrace:     []string{"main.main"},
				OutputTypeNames: []string{"*bytes.Buffer"},
			},
			wantMessage: "replaced",
			wantFields: map[string]any{
				"type":        "*bytes.Buffer",
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"module":      "myModule",
			},
		},
		{
			name: "Replace/Error",
			give: &fxevent.Replaced{
				StackTrace:  []string{"main.main", "runtime.main"},
				ModuleTrace: []string{"main.main"},
				Err:         someError,
			},

			wantMessage: "error encountered while replacing",
			wantFields: map[string]any{
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"error":       "some error",
			},
		},
		{
			name: "Decorate",
			give: &fxevent.Decorated{
				DecoratorName:   "bytes.NewBuffer()",
				StackTrace:      []string{"main.main", "runtime.main"},
				ModuleTrace:     []string{"main.main"},
				ModuleName:      "myModule",
				OutputTypeNames: []string{"*bytes.Buffer"},
			},
			wantMessage: "decorated",
			wantFields: map[string]any{
				"decorator":   "bytes.NewBuffer()",
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"type":        "*bytes.Buffer",
				"module":      "myModule",
			},
		},
		{
			name: "Decorate/Error",
			give: &fxevent.Decorated{
				StackTrace:  []string{"main.main", "runtime.main"},
				ModuleTrace: []string{"main.main"},
				Err:         someError,
			},
			wantMessage: "error encountered while applying options",
			wantFields: map[string]any{
				"stacktrace":  []any{"main.main", "runtime.main"},
				"moduletrace": []any{"main.main"},
				"error":       "some error",
			},
		},
		{
			name:        "Run",
			give:        &fxevent.Run{Name: "bytes.NewBuffer()", Kind: "constructor", Runtime: 3 * time.Millisecond},
			wantMessage: "run",
			wantFields: map[string]any{
				"name":    "bytes.NewBuffer()",
				"kind":    "constructor",
				"runtime": "3ms",
			},
		},
		{
			name: "Run with module",
			give: &fxevent.Run{
				Name:       "bytes.NewBuffer()",
				Kind:       "constructor",
				ModuleName: "myModule",
				Runtime:    3 * time.Millisecond,
			},
			wantMessage: "run",
			wantFields: map[string]any{
				"name":    "bytes.NewBuffer()",
				"kind":    "constructor",
				"module":  "myModule",
				"runtime": "3ms",
			},
		},
		{
			name: "Run/Error",
			give: &fxevent.Run{
				Name: "bytes.NewBuffer()",
				Kind: "constructor",
				Err:  someError,
			},
			wantMessage: "error returned",
			wantFields: map[string]any{
				"name":  "bytes.NewBuffer()",
				"kind":  "constructor",
				"error": "some error",
			},
		},
		{
			name:        "Invoking/Success",
			give:        &fxevent.Invoking{ModuleName: "myModule", FunctionName: "bytes.NewBuffer()"},
			wantMessage: "invoking",
			wantFields: map[string]any{
				"function": "bytes.NewBuffer()",
				"module":   "myModule",
			},
		},
		{
			name:        "Invoked/Error",
			give:        &fxevent.Invoked{FunctionName: "bytes.NewBuffer()", Err: someError},
			wantMessage: "invoke failed",
			wantFields: map[string]any{
				"error":    "some error",
				"stack":    "",
				"function": "bytes.NewBuffer()",
			},
		},
		{
			name:        "Start/Error",
			give:        &fxevent.Started{Err: someError},
			wantMessage: "start failed",
			wantFields: map[string]any{
				"error": "some error",
			},
		},
		{
			name:        "Stopping",
			give:        &fxevent.Stopping{Signal: os.Interrupt},
			wantMessage: "received signal",
			wantFields: map[string]any{
				"signal": "INTERRUPT",
			},
		},
		{
			name:        "Stopped/Error",
			give:        &fxevent.Stopped{Err: someError},
			wantMessage: "stop failed",
			wantFields: map[string]any{
				"error": "some error",
			},
		},
		{
			name:        "RollingBack/Error",
			give:        &fxevent.RollingBack{StartErr: someError},
			wantMessage: "start failed, rolling back",
			wantFields: map[string]any{
				"error": "some error",
			},
		},
		{
			name:        "RolledBack/Error",
			give:        &fxevent.RolledBack{Err: someError},
			wantMessage: "rollback failed",
			wantFields: map[string]any{
				"error": "some error",
			},
		},
		{
			name:        "Started",
			give:        &fxevent.Started{},
			wantMessage: "started",
			wantFields:  map[string]any{},
		},
		{
			name:        "LoggerInitialized/Error",
			give:        &fxevent.LoggerInitialized{Err: someError},
			wantMessage: "custom logger initialization failed",
			wantFields: map[string]any{
				"error": "some error",
			},
		},
		{
			name:        "LoggerInitialized",
			give:        &fxevent.LoggerInitialized{ConstructorName: "bytes.NewBuffer()"},
			wantMessage: "initialized custom fxevent.Logger",
			wantFields: map[string]any{
				"function": "bytes.NewBuffer()",
			},
		},
	}

	t.Run("debug observer, log at default (info)", func(t *testing.T) {
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				core, observedLogs := newZerologObservableLogger(zerolog.DebugLevel)
				(&ZerologLogger{Logger: core}).LogEvent(tt.give)

				logs := observedLogs.TakeAll()
				require.Len(t, logs, 1)
				got := logs[0]

				assert.Equal(t, tt.wantMessage, got.Message())
				assert.Equal(t, tt.wantFields, got.Fields())
			})
		}
	})

	t.Run("info observer, log at debug", func(t *testing.T) {
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				core, observedLogs := newZerologObservableLogger(zerolog.InfoLevel)
				l := &ZerologLogger{Logger: core}
				l.UseLogLevel(zerolog.DebugLevel)
				l.LogEvent(tt.give)

				logs := observedLogs.TakeAll()
				// logs are not visible unless they are errors
				if strings.HasSuffix(tt.name, "/Error") {
					require.Len(t, logs, 1)
					got := logs[0]
					assert.Equal(t, tt.wantMessage, got.Message())
					assert.Equal(t, tt.wantFields, got.Fields())
				} else {
					require.Len(t, logs, 0)
				}
			})
		}
	})

	t.Run("info observer, log/error at debug", func(t *testing.T) {
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				core, observedLogs := newZerologObservableLogger(zerolog.InfoLevel)
				l := &ZerologLogger{Logger: core}
				l.UseLogLevel(zerolog.DebugLevel)
				l.UseErrorLevel(zerolog.DebugLevel)
				l.LogEvent(tt.give)

				logs := observedLogs.TakeAll()
				require.Len(t, logs, 0, "no logs should be visible")
			})
		}
	})

	t.Run("test setting log levels", func(t *testing.T) {
		levels := []zerolog.Level{
			zerolog.ErrorLevel,
			zerolog.DebugLevel,
			zerolog.WarnLevel,
			zerolog.InfoLevel,
			zerolog.FatalLevel,
			zerolog.PanicLevel,
			zerolog.NoLevel,
			zerolog.Disabled,
			zerolog.TraceLevel,
		}

		for _, level := range levels {
			core, observedLogs := newZerologObservableLogger(level)
			logger := &ZerologLogger{Logger: core}
			logger.UseLogLevel(level)
			func() {
				defer func() {
					recover()
				}()
				logger.LogEvent(&fxevent.OnStartExecuting{
					FunctionName: "hook.onStart",
					CallerName:   "bytes.NewBuffer",
				})
			}()
			logs := observedLogs.TakeAll()
			if level == zerolog.Disabled {
				require.Len(t, logs, 0)
			} else {
				require.Len(t, logs, 1)
			}
		}
	})

	t.Run("test setting error log levels", func(t *testing.T) {
		levels := []zerolog.Level{
			zerolog.ErrorLevel,
			zerolog.DebugLevel,
			zerolog.WarnLevel,
			zerolog.InfoLevel,
			zerolog.FatalLevel,
			zerolog.PanicLevel,
			zerolog.NoLevel,
			zerolog.Disabled,
			zerolog.TraceLevel,
		}

		for _, level := range levels {
			core, observedLogs := newZerologObservableLogger(level)
			logger := &ZerologLogger{Logger: core}
			logger.UseErrorLevel(level)
			func() {
				defer func() {
					recover()
				}()
				logger.LogEvent(&fxevent.OnStopExecuted{
					FunctionName: "hook.onStart1",
					CallerName:   "bytes.NewBuffer",
					Err:          fmt.Errorf("some error"),
				})
			}()
			logs := observedLogs.TakeAll()
			if level == zerolog.Disabled {
				require.Len(t, logs, 0)
			} else {
				require.Len(t, logs, 1)
			}
		}
	})
}
