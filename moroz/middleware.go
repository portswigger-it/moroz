package moroz

import (
	"os"

	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log"
)

type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	logger = kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))

	return func(next Service) Service {
		return logmw{logger, next}
	}
}

type logmw struct {
	logger log.Logger
	next   Service
}
