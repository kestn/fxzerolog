# FxZerolog

[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/kestn/fxzerolog/master/LICENSE)
[![Build Status](https://github.com/rs/zerolog/actions/workflows/test.yml/badge.svg)](https://github.com/kestn/fxzerolog/actions/workflows/test.yml) [![codecov](https://codecov.io/github/kestn/fxzerolog/graph/badge.svg?token=1L60TMAFFJ)](https://codecov.io/github/kestn/fxzerolog)

An integration of [Zerolog](https://github.com/rs/zerolog) into [Uber Fx](https://github.com/uber-go/fx) framework for robust structured logging.

## Installation

```bash
go get -u github.com/kestn/fxzerolog
```

## Example

For simple logging, import the logger package **github.com/kestn/fxzerolog**

```go
package main

import (
  "github.com/kestn/fxzerolog"
  "github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
  "go.uber.org/fx"
  "go.uber.org/fx/fxevent"
)

func main() {
  fx.New(
    fx.Provide(func() zerolog.Logger { return log.Logger }),
    fx.WithLogger(func(log zerolog.Logger) fxevent.Logger {
      return &fxzerolog.ZerologLogger{Logger: log.With().Str("service", "fx").Logger()}
    }),
  ).Run()
}

// Output: {"level":"debug","service":"fx","constructor":"go.uber.org/fx.New.func1()","stacktrace":["go.uber.org/fx.New (/home/kestn/.cache/go/pkg/mod/go.uber.org/fx@v1.23.0/app.go:486)","main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:12)","runtime.main (/opt/go/src/runtime/proc.go:272)"],"moduletrace":["go.uber.org/fx.New (/home/kestn/.cache/go/pkg/mod/go.uber.org/fx@v1.23.0/app.go:486)","main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:12)"],"type":"fx.Lifecycle","time":"2024-12-30T00:41:42+01:00","message":"provided"}
// Output: {"level":"debug","service":"fx","constructor":"go.uber.org/fx.(*App).shutdowner-fm()","stacktrace":["go.uber.org/fx.New (/home/kestn/.cache/go/pkg/mod/go.uber.org/fx@v1.23.0/app.go:486)","main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:12)","runtime.main (/opt/go/src/runtime/proc.go:272)"],"moduletrace":["go.uber.org/fx.New (/home/kestn/.cache/go/pkg/mod/go.uber.org/fx@v1.23.0/app.go:486)","main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:12)"],"type":"fx.Shutdowner","time":"2024-12-30T00:41:42+01:00","message":"provided"}
// Output: {"level":"debug","service":"fx","constructor":"go.uber.org/fx.(*App).dotGraph-fm()","stacktrace":["go.uber.org/fx.New (/home/kestn/.cache/go/pkg/mod/go.uber.org/fx@v1.23.0/app.go:486)","main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:12)","runtime.main (/opt/go/src/runtime/proc.go:272)"],"moduletrace":["go.uber.org/fx.New (/home/kestn/.cache/go/pkg/mod/go.uber.org/fx@v1.23.0/app.go:486)","main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:12)"],"type":"fx.DotGraph","time":"2024-12-30T00:41:42+01:00","message":"provided"}
// Output: {"level":"debug","service":"fx","constructor":"main.main.func1()","stacktrace":["main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:13)","runtime.main (/opt/go/src/runtime/proc.go:272)"],"moduletrace":["main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:13)","main.main (/home/kestn/projects/github.com/kestn/fxzerolog/example/main.go:12)"],"type":"zerolog.Logger","time":"2024-12-30T00:41:42+01:00","message":"provided"}
// Output: {"level":"debug","service":"fx","name":"main.main.func1()","kind":"provide","runtime":"42.71µs","time":"2024-12-30T00:41:42+01:00","message":"run"}
// Output: {"level":"debug","service":"fx","function":"main.main.func2()","time":"2024-12-30T00:41:42+01:00","message":"initialized custom fxevent.Logger"}
// Output: {"level":"debug","service":"fx","time":"2024-12-30T00:41:42+01:00","message":"started"}
```

## Configuration

The `fxzerolog` project does not include any built-in configurations. 
However, you have the flexibility to customize the `zerolog.Logger` according to your specific requirements.

## License

FxZerolog is released under the MIT License. See [LICENSE](LICENSE)
