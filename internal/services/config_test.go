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

func TestConfigService_GetByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	expectedConfig := configs.Config{ID: 1, Name: "test"}

	mockQueries.EXPECT().
		GetConfigByName(gomock.Any(), "test").
		Return(expectedConfig, nil)

	config, err := service.GetByName(context.Background(), "test")

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("got %v, want %v", config, expectedConfig)
	}
}

func TestConfigService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	expectedConfigs := []configs.Config{
		{ID: 1, Name: "config1"},
		{ID: 2, Name: "config2"},
	}

	mockQueries.EXPECT().
		ListConfigs(gomock.Any(), gomock.Any()).
		Return(expectedConfigs, nil)

	configs, err := service.List(context.Background(), 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(configs, expectedConfigs) {
		t.Errorf("got %v, want %v", configs, expectedConfigs)
	}
}

func TestConfigService_ListByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	expectedConfigs := []configs.Config{
		{ID: 1, Name: "config1", Mode: "simple"},
	}

	mockQueries.EXPECT().
		ListConfigsByMode(gomock.Any(), gomock.Any()).
		Return(expectedConfigs, nil)

	configs, err := service.ListByMode(context.Background(), "simple", 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(configs, expectedConfigs) {
		t.Errorf("got %v, want %v", configs, expectedConfigs)
	}
}

func TestConfigService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	req := CreateConfigRequest{
		Name:        "updated-config",
		Description: "Updated config",
		Recipe: Recipe{
			Version: "1.0",
			Name:    "updated",
			Parameters: RecipeParameters{
				Alpha: 0.2,
			},
		},
		Tags: "updated-tag",
	}

	mockQueries.EXPECT().
		UpdateConfig(gomock.Any(), gomock.Any()).
		Return(nil)

	err := service.Update(context.Background(), 1, req)

	if err != nil {
		t.Fatal(err)
	}
}

func TestConfigService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	mockQueries.EXPECT().
		DeleteConfig(gomock.Any(), int64(1)).
		Return(nil)

	err := service.Delete(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
}

func TestConfigService_SetDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	config := configs.Config{ID: 1, Mode: "simple"}

	mockQueries.EXPECT().
		GetConfig(gomock.Any(), int64(1)).
		Return(config, nil)

	mockQueries.EXPECT().
		SetDefaultConfig(gomock.Any(), gomock.Any()).
		Return(nil)

	err := service.SetDefault(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
}

func TestConfigService_GetDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	expectedConfig := configs.Config{ID: 1, Name: "default", Mode: "simple"}

	mockQueries.EXPECT().
		GetDefaultConfig(gomock.Any(), "simple").
		Return(expectedConfig, nil)

	config, err := service.GetDefault(context.Background(), "simple")

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("got %v, want %v", config, expectedConfig)
	}
}

func TestConfigService_IncrementUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	mockQueries.EXPECT().
		IncrementConfigUsage(gomock.Any(), int64(1)).
		Return(nil)

	err := service.IncrementUsage(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
}

func TestConfigService_GetPreset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	expectedPreset := configs.ConfigPreset{ID: 1, Name: "preset1"}

	mockQueries.EXPECT().
		GetPreset(gomock.Any(), "preset1").
		Return(expectedPreset, nil)

	preset, err := service.GetPreset(context.Background(), "preset1")

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(preset, expectedPreset) {
		t.Errorf("got %v, want %v", preset, expectedPreset)
	}
}

func TestConfigService_validateRecipe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	// Valid recipe
	validRecipe := Recipe{
		Version: "1.0",
		Name:    "test",
		Parameters: RecipeParameters{
			Alpha: 0.1,
		},
	}

	err := service.validateRecipe(validRecipe)
	if err != nil {
		t.Errorf("expected no error for valid recipe, got %v", err)
	}

	// Invalid version
	invalidRecipe := Recipe{
		Version: "",
		Name:    "test",
		Parameters: RecipeParameters{
			Alpha: 0.1,
		},
	}

	err = service.validateRecipe(invalidRecipe)
	if err == nil {
		t.Error("expected error for invalid version, got nil")
	}

	// Invalid name
	invalidRecipe2 := Recipe{
		Version: "1.0",
		Name:    "",
		Parameters: RecipeParameters{
			Alpha: 0.1,
		},
	}

	err = service.validateRecipe(invalidRecipe2)
	if err == nil {
		t.Error("expected error for invalid name, got nil")
	}
}

func TestConfigService_validateRecipe_InvalidName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	invalidRecipe := Recipe{
		Version: "1.0",
		Name:    "", // Invalid
		Parameters: RecipeParameters{
			Alpha: 0.1,
		},
	}

	err := service.validateRecipe(invalidRecipe)
	if err == nil {
		t.Error("expected error for invalid name, got nil")
	}
}

func TestConfigService_Create_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	req := CreateConfigRequest{
		Name:        "test-config",
		Description: "Test config",
		Recipe: Recipe{
			Version: "", // Invalid
			Name:    "test",
			Parameters: RecipeParameters{
				Alpha: 0.1,
			},
		},
		Tags:      "tag1,tag2",
		Mode:      "simple",
		CreatedBy: "user1",
	}

	_, err := service.Create(context.Background(), req)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestConfigService_Update_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	req := CreateConfigRequest{
		Name:        "updated-config",
		Description: "Updated config",
		Recipe: Recipe{
			Version: "", // Invalid
			Name:    "updated",
			Parameters: RecipeParameters{
				Alpha: 0.2,
			},
		},
		Tags: "updated-tag",
	}

	err := service.Update(context.Background(), 1, req)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestConfigService_SetDefault_GetConfigError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	mockQueries.EXPECT().
		GetConfig(gomock.Any(), int64(1)).
		Return(configs.Config{}, sql.ErrNoRows)

	err := service.SetDefault(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConfigService_SetDefault_SetDefaultError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

	service := NewConfigService(mockQueries, nil, logger)

	config := configs.Config{ID: 1, Mode: "simple"}

	mockQueries.EXPECT().
		GetConfig(gomock.Any(), int64(1)).
		Return(config, nil)

	mockQueries.EXPECT().
		SetDefaultConfig(gomock.Any(), gomock.Any()).
		Return(sql.ErrConnDone)

	err := service.SetDefault(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConfigService_Create_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock.NewMockQuerier(ctrl)
	logger := &slog.Logger{}

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

	mockQueries.EXPECT().
		CreateConfig(gomock.Any(), gomock.Any()).
		Return(configs.Config{}, sql.ErrConnDone)

	_, err := service.Create(context.Background(), req)

	if err == nil {
		t.Fatal("expected database error, got nil")
	}
}
