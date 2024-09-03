// logging/logging.go
package logging

import (
    "os"
    "github.com/go-kit/log"
    "github.com/go-kit/log/level"
)

var Logger log.Logger

func InitLogger(debug bool) {
    Logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))

    if debug {
        Logger = level.NewFilter(Logger, level.AllowDebug())
    } else {
        Logger = level.NewFilter(Logger, level.AllowInfo())
    }

    Logger = log.With(Logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
}
