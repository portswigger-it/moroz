package moroz

import (
	"compress/zlib"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/groob/moroz/santa"
)

func (svc *SantaService) Preflight(ctx context.Context, machineID string, p santa.PreflightPayload) (*santa.Preflight, error) {
	config, err := svc.config(ctx, machineID)
	if err != nil {
		return nil, err
	}
	pre := config.Preflight
	return &pre, nil
}

type preflightRequest struct {
	MachineID string
	payload   santa.PreflightPayload
}

type preflightResponse struct {
	*santa.Preflight
	Err error `json:"error,omitempty"`
}

func (r preflightResponse) Failed() error { return r.Err }

func makePreflightEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(preflightRequest)
		preflight, err := svc.Preflight(ctx, req.MachineID, req.payload)
		if err != nil {
			return preflightResponse{Err: err}, nil
		}
		return preflightResponse{Preflight: preflight}, nil
	}
}

func decodePreflightRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	zr, err := zlib.NewReader(r.Body)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	defer r.Body.Close()
	id, err := machineIDFromRequest(r)
	if err != nil {
		return nil, err
	}
	//log.Printf("DEBUG: req body was %#v", zr)
	req := preflightRequest{MachineID: id}
	if err := json.NewDecoder(zr).Decode(&req.payload); err != nil {
		log.Printf("DEBUG: error was %#v", err)
		return nil, err
	}
	log.Printf("%#v", req)
	return req, nil
}

func (mw logmw) Preflight(ctx context.Context, machineID string, p santa.PreflightPayload) (pf *santa.Preflight, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "Preflight",
			"machine_id", machineID,
			"preflight_payload", p,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	pf, err = mw.next.Preflight(ctx, machineID, p)
	return
}
