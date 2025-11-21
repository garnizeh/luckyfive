package services

import (
	"context"
	"database/sql"
	"log/slog"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/garnizeh/luckyfive/internal/store/configs"
	"github.com/garnizeh/luckyfive/internal/store/configs/mock"
)

func TestConfigService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{} // or use a test logger

	service := NewConfigService(mockQueries, nil, logger)

	req := CreateConfigRequest{
		Name:        "test-config",
		Description: "Test config",
		Recipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: RecipeParameters{
				Alpha: 0.1,
			},
		},
		Tags:      "tag1,tag2",
		Mode:      "simple",
		CreatedBy: "user1",
	}

	expectedConfig := configs.Config{
		ID:          1,
		Name:        "test-config",
		Description: sql.NullString{String: "Test config", Valid: true},
		RecipeJson:  `{"version":"1.0","name":"test","parameters":{"alpha":0.1}}`,
		Tags:        sql.NullString{String: "tag1,tag2", Valid: true},
		IsDefault:   sql.NullInt64{Int64: 0, Valid: true},
		Mode:        "simple",
		CreatedBy:   sql.NullString{String: "user1", Valid: true},
	}

	mockQueries.EXPECT().
		CreateConfig(gomock.Any(), gomock.Any()).
		Return(expectedConfig, nil)

	config, err := service.Create(context.Background(), req)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("got %v, want %v", config, expectedConfig)
	}
}

func TestConfigService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	expectedConfig := configs.Config{ID: 1, Name: "test"}

	mockQueries.EXPECT().
		GetConfig(gomock.Any(), int64(1)).
		Return(expectedConfig, nil)

	config, err := service.Get(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("got %v, want %v", config, expectedConfig)
	}
}

func TestConfigService_ListPresets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	expectedPresets := []configs.ConfigPreset{
		{ID: 1, Name: "preset1"},
	}

	mockQueries.EXPECT().
		ListPresets(gomock.Any()).
		Return(expectedPresets, nil)

	presets, err := service.ListPresets(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(presets, expectedPresets) {
		t.Errorf("got %v, want %v", presets, expectedPresets)
	}
}
