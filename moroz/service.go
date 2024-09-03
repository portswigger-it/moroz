package moroz

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/groob/moroz/santa"
)

type ConfigStore interface {
	AllConfigs(ctx context.Context) ([]santa.Config, error)
	Config(ctx context.Context, machineID string) (santa.Config, error)
}

type SantaService struct {
	logger          log.Logger
	global          santa.Config
	repo            ConfigStore
	eventDir        string
	flPersistEvents bool
}

func NewService(ds ConfigStore, eventDir string, flPersistEvents bool, logger log.Logger) (*SantaService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	global, err := ds.Config(ctx, "global")
	if err != nil {
		logger.Log("level", "error", "msg", "Failed to fetch global config", "err", err)
		return nil, err
	}
	return &SantaService{
		logger:          logger,
		global:          global,
		repo:            ds,
		eventDir:        eventDir,
		flPersistEvents: flPersistEvents,
	}, nil
}

type Service interface {
	Preflight(ctx context.Context, machineID string, p santa.PreflightPayload) (*santa.Preflight, error)
	RuleDownload(ctx context.Context, machineID string) ([]santa.Rule, error)
	UploadEvent(ctx context.Context, machineID string, events []santa.EventPayload) error
}

type Endpoints struct {
	PreflightEndpoint    endpoint.Endpoint
	RuleDownloadEndpoint endpoint.Endpoint
	EventUploadEndpoint  endpoint.Endpoint
}

func MakeServerEndpoints(svc Service) Endpoints {
	return Endpoints{
		PreflightEndpoint:    makePreflightEndpoint(svc),
		RuleDownloadEndpoint: makeRuleDownloadEndpoint(svc),
		EventUploadEndpoint:  makeEventUploadEndpoint(svc),
	}
}
