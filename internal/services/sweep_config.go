package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/garnizeh/luckyfive/internal/store/sweeps"
	"github.com/garnizeh/luckyfive/pkg/sweep"
)

type SweepConfigService struct {
	sweepsQueries sweeps.Querier
	sweepsDB      *sql.DB
	logger        *slog.Logger
}

type SweepConfigServicer interface {
	Create(ctx context.Context, req CreateSweepConfigRequest) (sweeps.Sweep, error)
	Get(ctx context.Context, id int64) (sweeps.Sweep, error)
	GetByName(ctx context.Context, name string) (sweeps.Sweep, error)
	List(ctx context.Context, limit, offset int64) ([]sweeps.Sweep, error)
	Update(ctx context.Context, id int64, req CreateSweepConfigRequest) error
	Delete(ctx context.Context, id int64) error
	IncrementUsage(ctx context.Context, id int64) error
}

func NewSweepConfigService(
	sweepsQueries sweeps.Querier,
	sweepsDB *sql.DB,
	logger *slog.Logger,
) *SweepConfigService {
	return &SweepConfigService{
		sweepsQueries: sweepsQueries,
		sweepsDB:      sweepsDB,
		logger:        logger,
	}
}

type SweepConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Config      sweep.SweepConfig `json:"config"`
	CreatedBy   string            `json:"created_by"`
}

type CreateSweepConfigRequest struct {
	Name        string
	Description string
	Config      sweep.SweepConfig
	CreatedBy   string
}

func (s *SweepConfigService) Create(ctx context.Context, req CreateSweepConfigRequest) (sweeps.Sweep, error) {
	// Validate sweep config
	if err := req.Config.Validate(); err != nil {
		return sweeps.Sweep{}, fmt.Errorf("invalid sweep config: %w", err)
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return sweeps.Sweep{}, fmt.Errorf("marshal config: %w", err)
	}

	// Create sweep record
	sweepRecord, err := s.sweepsQueries.CreateSweep(ctx, sweeps.CreateSweepParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		ConfigJson:  string(configJSON),
		CreatedBy:   sql.NullString{String: req.CreatedBy, Valid: req.CreatedBy != ""},
	})
	if err != nil {
		return sweeps.Sweep{}, fmt.Errorf("create sweep: %w", err)
	}

	s.logger.Info("created sweep config", "id", sweepRecord.ID, "name", sweepRecord.Name)
	return sweepRecord, nil
}

func (s *SweepConfigService) Get(ctx context.Context, id int64) (sweeps.Sweep, error) {
	return s.sweepsQueries.GetSweep(ctx, id)
}

func (s *SweepConfigService) GetByName(ctx context.Context, name string) (sweeps.Sweep, error) {
	return s.sweepsQueries.GetSweepByName(ctx, name)
}

func (s *SweepConfigService) List(ctx context.Context, limit, offset int64) ([]sweeps.Sweep, error) {
	return s.sweepsQueries.ListSweeps(ctx, sweeps.ListSweepsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (s *SweepConfigService) Update(ctx context.Context, id int64, req CreateSweepConfigRequest) error {
	// Validate sweep config
	if err := req.Config.Validate(); err != nil {
		return fmt.Errorf("invalid sweep config: %w", err)
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return s.sweepsQueries.UpdateSweep(ctx, sweeps.UpdateSweepParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		ConfigJson:  string(configJSON),
		ID:          id,
	})
}

func (s *SweepConfigService) Delete(ctx context.Context, id int64) error {
	return s.sweepsQueries.DeleteSweep(ctx, id)
}

func (s *SweepConfigService) IncrementUsage(ctx context.Context, id int64) error {
	return s.sweepsQueries.IncrementSweepUsage(ctx, id)
}
