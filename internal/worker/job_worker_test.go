package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	servicemock "github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
	storemock "github.com/garnizeh/luckyfive/internal/store/simulations/mock"
	"github.com/golang/mock/gomock"
)

func TestJobWorker_Start_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSimSvc := servicemock.NewMockSimulationServicer(ctrl)
	mockQuerier := storemock.NewMockQuerier(ctrl)

	// Expect no calls since no jobs are available
	mockQuerier.EXPECT().ClaimPendingSimulation(gomock.Any(), gomock.Any()).Return(simulations.Simulation{}, sql.ErrNoRows).AnyTimes()

	// Create a discard logger for testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	worker := NewJobWorker(
		mockQuerier,
		mockSimSvc,
		"test-worker",
		100*time.Millisecond, // Very short poll interval for testing
		2,
		logger, // Use proper logger
	)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start worker in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Start(ctx)
	}()

	// Wait a bit then stop
	time.Sleep(200 * time.Millisecond)
	worker.Stop()

	// Wait for worker to stop
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Errorf("Expected no error or context canceled, got %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Worker did not stop within timeout")
	}
}

func TestJobWorker_GracefulShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSimSvc := servicemock.NewMockSimulationServicer(ctrl)
	mockQuerier := storemock.NewMockQuerier(ctrl)

	// Expect no calls since no jobs are available
	mockQuerier.EXPECT().ClaimPendingSimulation(gomock.Any(), gomock.Any()).Return(simulations.Simulation{}, sql.ErrNoRows).AnyTimes()

	// Create a discard logger for testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	worker := NewJobWorker(
		mockQuerier,
		mockSimSvc,
		"test-worker",
		1*time.Second, // Long poll interval
		1,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())

	// Start worker
	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context (simulates SIGINT)
	cancel()

	// Wait for worker to stop gracefully
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Errorf("Expected context canceled error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Worker did not stop gracefully within timeout")
	}
}
