package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/garnizeh/luckyfive/internal/store/configs"
)

type ConfigService struct {
	configsQueries configs.Querier
	configsDB      *sql.DB
	logger         *slog.Logger
}

func NewConfigService(
	configsQueries configs.Querier,
	configsDB *sql.DB,
	logger *slog.Logger,
) *ConfigService {
	return &ConfigService{
		configsQueries: configsQueries,
		configsDB:      configsDB,
		logger:         logger,
	}
}

type Config struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Recipe      Recipe `json:"recipe"`
	Tags        string `json:"tags"`
	IsDefault   bool   `json:"is_default"`
	Mode        string `json:"mode"`
	CreatedBy   string `json:"created_by"`
}

type CreateConfigRequest struct {
	Name        string
	Description string
	Recipe      Recipe
	Tags        string
	Mode        string
	CreatedBy   string
}

func (s *ConfigService) Create(ctx context.Context, req CreateConfigRequest) (configs.Config, error) {
	// Validate recipe
	if err := s.validateRecipe(req.Recipe); err != nil {
		return configs.Config{}, fmt.Errorf("invalid recipe: %w", err)
	}

	// Marshal recipe to JSON
	recipeJSON, err := json.Marshal(req.Recipe)
	if err != nil {
		return configs.Config{}, fmt.Errorf("marshal recipe: %w", err)
	}

	// Create config record
	config, err := s.configsQueries.CreateConfig(ctx, configs.CreateConfigParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		RecipeJson:  string(recipeJSON),
		Tags:        sql.NullString{String: req.Tags, Valid: req.Tags != ""},
		IsDefault:   sql.NullInt64{Int64: 0, Valid: true}, // New configs are not default
		Mode:        req.Mode,
		CreatedBy:   sql.NullString{String: req.CreatedBy, Valid: req.CreatedBy != ""},
	})
	if err != nil {
		return configs.Config{}, fmt.Errorf("create config: %w", err)
	}

	return config, nil
}

func (s *ConfigService) Get(ctx context.Context, id int64) (configs.Config, error) {
	return s.configsQueries.GetConfig(ctx, id)
}

func (s *ConfigService) GetByName(ctx context.Context, name string) (configs.Config, error) {
	return s.configsQueries.GetConfigByName(ctx, name)
}

func (s *ConfigService) List(ctx context.Context, limit, offset int64) ([]configs.Config, error) {
	return s.configsQueries.ListConfigs(ctx, configs.ListConfigsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (s *ConfigService) ListByMode(ctx context.Context, mode string, limit, offset int64) ([]configs.Config, error) {
	return s.configsQueries.ListConfigsByMode(ctx, configs.ListConfigsByModeParams{
		Mode:   mode,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *ConfigService) Update(ctx context.Context, id int64, req CreateConfigRequest) error {
	// Validate recipe
	if err := s.validateRecipe(req.Recipe); err != nil {
		return fmt.Errorf("invalid recipe: %w", err)
	}

	// Marshal recipe to JSON
	recipeJSON, err := json.Marshal(req.Recipe)
	if err != nil {
		return fmt.Errorf("marshal recipe: %w", err)
	}

	return s.configsQueries.UpdateConfig(ctx, configs.UpdateConfigParams{
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		RecipeJson:  string(recipeJSON),
		Tags:        sql.NullString{String: req.Tags, Valid: req.Tags != ""},
		ID:          id,
	})
}

func (s *ConfigService) Delete(ctx context.Context, id int64) error {
	return s.configsQueries.DeleteConfig(ctx, id)
}

func (s *ConfigService) SetDefault(ctx context.Context, id int64) error {
	// Get the config to find its mode
	config, err := s.configsQueries.GetConfig(ctx, id)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	return s.configsQueries.SetDefaultConfig(ctx, configs.SetDefaultConfigParams{
		ID:   id,
		Mode: config.Mode,
	})
}

func (s *ConfigService) GetDefault(ctx context.Context, mode string) (configs.Config, error) {
	return s.configsQueries.GetDefaultConfig(ctx, mode)
}

func (s *ConfigService) IncrementUsage(ctx context.Context, id int64) error {
	return s.configsQueries.IncrementConfigUsage(ctx, id)
}

func (s *ConfigService) GetPreset(ctx context.Context, name string) (configs.ConfigPreset, error) {
	return s.configsQueries.GetPreset(ctx, name)
}

func (s *ConfigService) ListPresets(ctx context.Context) ([]configs.ConfigPreset, error) {
	return s.configsQueries.ListPresets(ctx)
}

func (s *ConfigService) validateRecipe(recipe Recipe) error {
	if recipe.Version == "" {
		return fmt.Errorf("recipe version is required")
	}
	if recipe.Name == "" {
		return fmt.Errorf("recipe name is required")
	}
	// Add more validation as needed
	return nil
}
