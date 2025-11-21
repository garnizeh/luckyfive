package services

import (
	"context"
	"database/sql"
	"log/slog"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/garnizeh/luckyfive/internal/store/simulations"
	"github.com/garnizeh/luckyfive/internal/store/simulations/mock"
)

func TestSimulationService_CreateSimulation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	req := CreateSimulationRequest{
		Mode:       "simple",
		RecipeName: "test-recipe",
		Recipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: RecipeParameters{
				Alpha:      0.1,
				Beta:       0.2,
				Gamma:      0.3,
				Delta:      0.4,
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
		StartContest: 1000,
		EndContest:   1010,
		Async:        true,
		CreatedBy:    "user1",
	}

	expectedSim := simulations.Simulation{
		ID:           1,
		RecipeName:   sql.NullString{String: "test-recipe", Valid: true},
		RecipeJson:   `{"version":"1.0","name":"test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5,"enableEvolutionary":false,"generations":0,"mutationRate":0}}`,
		Mode:         "simple",
		StartContest: 1000,
		EndContest:   1010,
		Status:       "pending",
		CreatedBy:    sql.NullString{String: "user1", Valid: true},
	}

	mockQueries.EXPECT().
		CreateSimulation(gomock.Any(), gomock.Any()).
		Return(expectedSim, nil)

	sim, err := service.CreateSimulation(context.Background(), req)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*sim, expectedSim) {
		t.Errorf("got %v, want %v", *sim, expectedSim)
	}
}

func TestSimulationService_GetSimulation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	expectedSim := simulations.Simulation{
		ID:     1,
		Status: "completed",
	}

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(expectedSim, nil)

	sim, err := service.GetSimulation(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*sim, expectedSim) {
		t.Errorf("got %v, want %v", *sim, expectedSim)
	}
}

func TestSimulationService_CancelSimulation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	mockQueries.EXPECT().
		CancelSimulation(gomock.Any(), gomock.Any()).
		Return(nil)

	err := service.CancelSimulation(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
}

func TestSimulationService_ListSimulations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	expectedSims := []simulations.Simulation{
		{ID: 1, Status: "pending"},
		{ID: 2, Status: "running"},
	}

	mockQueries.EXPECT().
		ListSimulations(gomock.Any(), gomock.Any()).
		Return(expectedSims, nil)

	sims, err := service.ListSimulations(context.Background(), 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sims, expectedSims) {
		t.Errorf("got %v, want %v", sims, expectedSims)
	}
}

func TestSimulationService_ListSimulationsByStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	expectedSims := []simulations.Simulation{
		{ID: 1, Status: "pending"},
	}

	mockQueries.EXPECT().
		ListSimulationsByStatus(gomock.Any(), gomock.Any()).
		Return(expectedSims, nil)

	sims, err := service.ListSimulationsByStatus(context.Background(), "pending", 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sims, expectedSims) {
		t.Errorf("got %v, want %v", sims, expectedSims)
	}
}
