module github.com/kestn/fxzerolog/example

go 1.23.4

replace github.com/kestn/fxzerolog v0.0.1 => ../

require (
	github.com/kestn/fxzerolog v0.0.1
	github.com/rs/zerolog v1.33.0
	go.uber.org/fx v1.23.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
