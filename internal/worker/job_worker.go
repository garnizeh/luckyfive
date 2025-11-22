package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

type JobWorker struct {
	simulationsQueries simulations.Querier
	simulationService  services.SimulationServicer
	workerID           string
	pollInterval       time.Duration
	maxConcurrent      int
	logger             *slog.Logger
	shutdown           chan struct{}
}

func NewJobWorker(
	simulationsQueries simulations.Querier,
	simulationService services.SimulationServicer,
	workerID string,
	pollInterval time.Duration,
	maxConcurrent int,
	logger *slog.Logger,
) *JobWorker {
	return &JobWorker{
		simulationsQueries: simulationsQueries,
		simulationService:  simulationService,
		workerID:           workerID,
		pollInterval:       pollInterval,
		maxConcurrent:      maxConcurrent,
		logger:             logger,
		shutdown:           make(chan struct{}),
	}
}

func (w *JobWorker) Start(ctx context.Context) error {
	w.logger.Info("worker starting", "worker_id", w.workerID)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Semaphore for concurrency control
	sem := make(chan struct{}, w.maxConcurrent)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker shutting down", "worker_id", w.workerID)
			return ctx.Err()
		case <-w.shutdown:
			w.logger.Info("worker stopped", "worker_id", w.workerID)
			return nil
		case <-ticker.C:
			// Try to claim a job
			job, err := w.simulationsQueries.ClaimPendingSimulation(ctx, simulations.ClaimPendingSimulationParams{
				StartedAt: sql.NullString{String: time.Now().Format(time.RFC3339), Valid: true},
				WorkerID:  sql.NullString{String: w.workerID, Valid: true},
			})
			if err != nil {
				if err != sql.ErrNoRows {
					w.logger.Error("claim job failed", "error", err, "worker_id", w.workerID)
				}
				continue
			}

			// Execute in goroutine
			sem <- struct{}{}
			go func(jobID int64) {
				defer func() { <-sem }()

				w.logger.Info("processing job", "job_id", jobID, "worker_id", w.workerID)

				if err := w.simulationService.ExecuteSimulation(ctx, jobID); err != nil {
					w.logger.Error("job execution failed", "job_id", jobID, "error", err, "worker_id", w.workerID)
				} else {
					w.logger.Info("job completed", "job_id", jobID, "worker_id", w.workerID)
				}
			}(job.ID)
		}
	}
}

func (w *JobWorker) Stop() {
	close(w.shutdown)
}
