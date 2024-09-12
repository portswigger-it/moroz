package moroz

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/groob/moroz/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func AddHTTPRoutes(r *mux.Router, e Endpoints, logger log.Logger) {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerAfter(
			httptransport.SetContentType("application/json; charset=utf-8"),
		),
	}

	// POST     /v1/santa/preflight/:id			preflight request.
	// POST     /v1/santa/ruledownload/:id		request rule updates.
	// POST     /v1/santa/eventupload/:id		upload event.
	// POST     /v1/santa/postflight/:id		postflight request. Implemented as a no-op.

	// Preflight Route
	r.Methods("POST").Path("/v1/santa/preflight/{id}").Handler(httptransport.NewServer(
		e.PreflightEndpoint,
		decodePreflightRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(func(ctx context.Context, req *http.Request) context.Context {
			metrics.PreflightRequests.WithLabelValues("POST").Inc() // Increment preflight requests counter
			return ctx
		}))...,
	))

	// Rule Download Route
	r.Methods("POST").Path("/v1/santa/ruledownload/{id}").Handler(httptransport.NewServer(
		e.RuleDownloadEndpoint,
		decodeRuleRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(func(ctx context.Context, req *http.Request) context.Context {
			metrics.RuleDownloadRequests.WithLabelValues("POST").Inc() // Increment rule download requests counter
			return ctx
		}))...,
	))

	// Event Upload Route
	r.Methods("POST").Path("/v1/santa/eventupload/{id}").Handler(httptransport.NewServer(
		e.EventUploadEndpoint,
		decodeEventUpload,
		encodeResponse,
		append(options, httptransport.ServerBefore(func(ctx context.Context, req *http.Request) context.Context {
			metrics.EventUploadRequests.WithLabelValues("POST").Inc() // Increment event upload requests counter
			return ctx
		}))...,
	))

	// Postflight Route (no-op)
	r.Methods("POST").Path("/v1/santa/postflight/{id}").Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			metrics.PostflightRequests.WithLabelValues("POST").Inc() // Increment postflight requests counter
		},
	))

	// Health Check Route
	r.Methods("GET").Path("/healthz").Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))

	// Add Prometheus
	r.Methods("GET").Path("/metrics").Handler(promhttp.Handler())
}

// errBadRoute is used for mux errors
var errBadRoute = errors.New("bad route")

func machineIDFromRequest(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return "", errBadRoute
	}
	return id, nil
}

type failer interface {
	Failed() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	if headerer, ok := response.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}
	code := http.StatusOK
	if sc, ok := response.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(response)
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	errMap := map[string]interface{}{"error": err.Error()}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if headerer, ok := err.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}

	code := http.StatusInternalServerError
	if sc, ok := err.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)

	enc.Encode(errMap)
}
