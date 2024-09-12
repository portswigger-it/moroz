package moroz

import (
	"compress/zlib"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"

	"github.com/groob/moroz/metrics"
	"github.com/groob/moroz/santa"
)

/*
func (svc *SantaService) UploadEvent(ctx context.Context, machineID string, events []santa.EventPayload) error {
	// TODO
	if !svc.flPersistEvents {
		return nil
	}
	for _, ev := range events {
		eventDir := filepath.Join(svc.eventDir, ev.FileSHA, machineID)
		if err := os.MkdirAll(eventDir, 0700); err != nil {
			svc.logger.Log("level", "error", "msg", "Failed to create event directory", "eventDir", eventDir, "err", err)
			return errors.Wrapf(err, "create event directory %s", eventDir)
		}

		eventPath := filepath.Join(eventDir, fmt.Sprintf("%f.json", ev.UnixTime))
		ev.EventInfo.MachineID = machineID

		eventInfoJSON, err := json.Marshal(ev.EventInfo)
		if err != nil {
			svc.logger.Log("level", "error", "msg", "Failed to marshal event info to JSON", "err", err)
			return errors.Wrap(err, "marshal event info to json")
		}
		if err := os.WriteFile(eventPath, eventInfoJSON, 0644); err != nil {
			svc.logger.Log("level", "error", "msg", "Failed to write event to path", "eventPath", eventPath, "err", err)
			return errors.Wrapf(err, "write event to path %s", eventPath)
		}
		svc.logger.Log(
			"event", "UploadEvent",
			"machineID", machineID,
			"eventInfo", string(eventInfoJSON),
			"eventPath", eventPath,
		)
	}
	return nil
}
*/

func (svc *SantaService) UploadEvent(ctx context.Context, machineID string, events []santa.EventPayload) error {
	// Increment the counter for the number of events processed
	metrics.EventProcessedCount.WithLabelValues(machineID).Add(float64(len(events)))

	for _, ev := range events {
		ev.EventInfo.MachineID = machineID

		// Track the decision outcome for each event
		metrics.DecisionOutcomes.WithLabelValues(ev.EventInfo.Decision).Inc()

		// Marshal the event info to JSON for logging purposes
		eventInfoJSON, err := json.Marshal(ev.EventInfo)
		if err != nil {
			svc.logger.Log("level", "error", "msg", "Failed to marshal event info to JSON", "err", err)
			// Increment the error counter for marshaling errors
			metrics.EventMarshalingErrors.WithLabelValues(machineID).Inc()
			return errors.Wrap(err, "marshal event info to json")
		}

		// Log the event information instead of writing it to a file
		svc.logger.Log(
			"event", "UploadEvent",
			"machineID", machineID,
			"eventInfo", string(eventInfoJSON),
		)
	}

	return nil
}

type eventRequest struct {
	MachineID string
	events    []santa.EventPayload
}

type eventResponse struct {
	Err error
}

func (r eventResponse) Failed() error { return r.Err }

func makeEventUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(eventRequest)
		err := svc.UploadEvent(ctx, req.MachineID, req.events)
		return eventResponse{Err: err}, nil
	}
}

func decodeEventUpload(ctx context.Context, r *http.Request) (interface{}, error) {
	// santa sends zlib compressed payloads
	zr, err := zlib.NewReader(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "create zlib reader to decode event upload")
	}
	defer zr.Close()

	id, err := machineIDFromRequest(r)
	if err != nil {
		return nil, errors.Wrap(err, "get machine ID from event upload URL")
	}

	// decode the JSON into individual log events.
	var eventPayload santa.EventUploadRequest

	if err := json.NewDecoder(zr).Decode(&eventPayload); err != nil {
		return nil, errors.Wrap(err, "decoding event upload request json")
	}

	var events []santa.EventPayload
	for _, ev := range eventPayload.Events {
		var payload santa.EventPayload
		payload.EventInfo = ev
		payload.FileSHA = ev.FileSHA256
		payload.UnixTime = ev.ExecutionTime
		events = append(events, payload)
	}

	req := eventRequest{MachineID: id, events: events}
	return req, nil
}

func (mw logmw) UploadEvent(ctx context.Context, machineID string, events []santa.EventPayload) (err error) {
	defer func(begin time.Time) {
		status := "success"
		if err != nil {
			status = "error"
		}

		// Increment the event upload request counter
		metrics.EventUploadRequests.WithLabelValues("POST").Inc()

		// Observe the request duration using time.Since(begin)
		metrics.EventUploadRequestDuration.WithLabelValues(status).Observe(time.Since(begin).Seconds())

		for _, ev := range events {
			_ = mw.logger.Log(
				"method", "UploadEvent",
				"machine_id", machineID,
				"event", ev.EventInfo,
				"err", err,
				"took", time.Since(begin),
			)
		}
	}(time.Now())

	err = mw.next.UploadEvent(ctx, machineID, events)
	return
}
