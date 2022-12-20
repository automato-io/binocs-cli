package main

import (
	"time"

	"github.com/automato-io/binocs-cli/cmd"
	"github.com/getsentry/sentry-go"
)

const (
	sentryDsn = "https://3d0ba3056a3f40fd8c743eb9e8234cb3@o4504359623196672.ingest.sentry.io/4504359625752576"
)

func main() {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		SampleRate:       1.0,
		TracesSampleRate: 1.0,
		Debug:            false,
		Release:          cmd.BinocsVersion,
	})
	defer sentry.Flush(10 * time.Second)

	cmd.Execute()
}
