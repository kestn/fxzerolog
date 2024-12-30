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
