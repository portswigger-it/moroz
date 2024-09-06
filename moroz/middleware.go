package moroz

import (
	"github.com/go-kit/log"
	"github.com/groob/moroz/logging"
)

type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return logmw{logger: logging.Logger, next: next}
	}
}

type logmw struct {
	logger log.Logger
	next   Service
}
