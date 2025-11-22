package services

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/garnizeh/luckyfive/internal/store/sweeps"
	"github.com/garnizeh/luckyfive/internal/store/sweeps/mock"
	"github.com/garnizeh/luckyfive/pkg/sweep"
)

func TestSweepConfigService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	req := CreateSweepConfigRequest{
		Name:        "test_sweep",
		Description: "Test sweep",
		Config: sweep.SweepConfig{
			Name:        "test_sweep",
			Description: "Test sweep configuration",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "advanced",
				Parameters: map[string]any{
					"alpha": 0.1,
				},
			},
			Parameters: []sweep.ParameterSweep{
				{
					Name: "alpha",
					Type: "range",
					Values: sweep.RangeValues{
						Min:  0.0,
						Max:  1.0,
						Step: 0.1,
					},
				},
			},
		},
		CreatedBy: "test_user",
	}

	expectedSweep := sweeps.Sweep{
		ID:          1,
		Name:        "test_sweep",
		Description: sql.NullString{String: "Test sweep", Valid: true},
		ConfigJson:  `{"name":"test_sweep","description":"Test sweep configuration","base_recipe":{"version":"1.0","name":"advanced","parameters":{"alpha":0.1}},"parameters":[{"name":"alpha","type":"range","values":{"min":0,"max":1,"step":0.1}}]}`,
		CreatedAt:   "2025-01-01 00:00:00",
		UpdatedAt:   "2025-01-01 00:00:00",
		CreatedBy:   sql.NullString{String: "test_user", Valid: true},
		TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
		LastUsedAt:  sql.NullString{Valid: false},
	}

	mockQueries.EXPECT().
		CreateSweep(gomock.Any(), gomock.Any()).
		Return(expectedSweep, nil)

	result, err := service.Create(context.Background(), req)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expectedSweep) {
		t.Errorf("got %+v, want %+v", result, expectedSweep)
	}
}

func TestSweepConfigService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	expectedSweep := sweeps.Sweep{
		ID:          1,
		Name:        "test_sweep",
		Description: sql.NullString{String: "Test sweep", Valid: true},
		ConfigJson:  `{"name":"test_sweep"}`,
		CreatedAt:   "2025-01-01 00:00:00",
		UpdatedAt:   "2025-01-01 00:00:00",
		CreatedBy:   sql.NullString{String: "test_user", Valid: true},
		TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
		LastUsedAt:  sql.NullString{Valid: false},
	}

	mockQueries.EXPECT().
		GetSweep(gomock.Any(), int64(1)).
		Return(expectedSweep, nil)

	result, err := service.Get(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expectedSweep) {
		t.Errorf("got %+v, want %+v", result, expectedSweep)
	}
}

func TestSweepConfigService_GetByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	expectedSweep := sweeps.Sweep{
		ID:          1,
		Name:        "test_sweep",
		Description: sql.NullString{String: "Test sweep", Valid: true},
		ConfigJson:  `{"name":"test_sweep"}`,
		CreatedAt:   "2025-01-01 00:00:00",
		UpdatedAt:   "2025-01-01 00:00:00",
		CreatedBy:   sql.NullString{String: "test_user", Valid: true},
		TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
		LastUsedAt:  sql.NullString{Valid: false},
	}

	mockQueries.EXPECT().
		GetSweepByName(gomock.Any(), "test_sweep").
		Return(expectedSweep, nil)

	result, err := service.GetByName(context.Background(), "test_sweep")

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expectedSweep) {
		t.Errorf("got %+v, want %+v", result, expectedSweep)
	}
}

func TestSweepConfigService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	expectedSweeps := []sweeps.Sweep{
		{
			ID:          1,
			Name:        "test_sweep_1",
			Description: sql.NullString{String: "Test sweep 1", Valid: true},
			ConfigJson:  `{"name":"test_sweep_1"}`,
			CreatedAt:   "2025-01-01 00:00:00",
			UpdatedAt:   "2025-01-01 00:00:00",
			CreatedBy:   sql.NullString{String: "test_user", Valid: true},
			TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
			LastUsedAt:  sql.NullString{Valid: false},
		},
		{
			ID:          2,
			Name:        "test_sweep_2",
			Description: sql.NullString{String: "Test sweep 2", Valid: true},
			ConfigJson:  `{"name":"test_sweep_2"}`,
			CreatedAt:   "2025-01-02 00:00:00",
			UpdatedAt:   "2025-01-02 00:00:00",
			CreatedBy:   sql.NullString{String: "test_user", Valid: true},
			TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
			LastUsedAt:  sql.NullString{Valid: false},
		},
	}

	mockQueries.EXPECT().
		ListSweeps(gomock.Any(), gomock.Any()).
		Return(expectedSweeps, nil)

	result, err := service.List(context.Background(), 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expectedSweeps) {
		t.Errorf("got %+v, want %+v", result, expectedSweeps)
	}
}

func TestSweepConfigService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	req := CreateSweepConfigRequest{
		Name:        "updated_sweep",
		Description: "Updated sweep",
		Config: sweep.SweepConfig{
			Name:        "updated_sweep",
			Description: "Updated sweep configuration",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "advanced",
				Parameters: map[string]any{
					"alpha": 0.2,
				},
			},
			Parameters: []sweep.ParameterSweep{
				{
					Name: "alpha",
					Type: "range",
					Values: sweep.RangeValues{
						Min:  0.0,
						Max:  1.0,
						Step: 0.1,
					},
				},
			},
		},
		CreatedBy: "test_user",
	}

	mockQueries.EXPECT().
		UpdateSweep(gomock.Any(), gomock.Any()).
		Return(nil)

	err := service.Update(context.Background(), 1, req)

	if err != nil {
		t.Fatal(err)
	}
}

func TestSweepConfigService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	mockQueries.EXPECT().
		DeleteSweep(gomock.Any(), int64(1)).
		Return(nil)

	err := service.Delete(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
}

func TestSweepConfigService_IncrementUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	mockQueries.EXPECT().
		IncrementSweepUsage(gomock.Any(), int64(1)).
		Return(nil)

	err := service.IncrementUsage(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
}

func TestSweepConfigService_Create_InvalidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	// Invalid config - missing name
	req := CreateSweepConfigRequest{
		Description: "Test sweep",
		Config: sweep.SweepConfig{
			Description: "Test sweep configuration",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "advanced",
				Parameters: map[string]any{
					"alpha": 0.1,
				},
			},
			Parameters: []sweep.ParameterSweep{
				{
					Name: "alpha",
					Type: "range",
					Values: sweep.RangeValues{
						Min:  0.0,
						Max:  1.0,
						Step: 0.1,
					},
				},
			},
		},
		CreatedBy: "test_user",
	}

	_, err := service.Create(context.Background(), req)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestSweepConfigService_Update_InvalidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	service := NewSweepConfigService(mockQueries, nil, logger)

	// Invalid config - missing name
	req := CreateSweepConfigRequest{
		Description: "Test sweep",
		Config: sweep.SweepConfig{
			Description: "Test sweep configuration",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "advanced",
				Parameters: map[string]any{
					"alpha": 0.1,
				},
			},
			Parameters: []sweep.ParameterSweep{
				{
					Name: "alpha",
					Type: "range",
					Values: sweep.RangeValues{
						Min:  0.0,
						Max:  1.0,
						Step: 0.1,
					},
				},
			},
		},
		CreatedBy: "test_user",
	}

	err := service.Update(context.Background(), 1, req)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
