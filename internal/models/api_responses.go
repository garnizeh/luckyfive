package models

// API response types for swagger documentation
// These types avoid sql.Null* types that swagger can't handle

// SweepJobResponse represents a sweep job for API responses
type SweepJobResponse struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	SweepConfigJson      string `json:"sweep_config_json"`
	BaseContestRange     string `json:"base_contest_range"`
	Status               string `json:"status"`
	TotalCombinations    int64  `json:"total_combinations"`
	CompletedSimulations int64  `json:"completed_simulations,omitempty"`
	FailedSimulations    int64  `json:"failed_simulations,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	StartedAt            string `json:"started_at,omitempty"`
	FinishedAt           string `json:"finished_at,omitempty"`
	RunDurationMs        int64  `json:"run_duration_ms,omitempty"`
	CreatedBy            string `json:"created_by,omitempty"`
}

// SweepStatusResponse represents sweep status for API responses
type SweepStatusResponse struct {
	Sweep       SweepJobResponse                `json:"sweep"`
	Total       int                             `json:"total"`
	Completed   int                             `json:"completed"`
	Running     int                             `json:"running"`
	Failed      int                             `json:"failed"`
	Pending     int                             `json:"pending"`
	Simulations []SweepSimulationDetailResponse `json:"simulations"`
}

// SweepSimulationDetailResponse represents simulation details in sweep status
type SweepSimulationDetailResponse struct {
	ID              int64  `json:"id"`
	SweepJobID      int64  `json:"sweep_job_id"`
	SimulationID    int64  `json:"simulation_id"`
	VariationIndex  int64  `json:"variation_index"`
	VariationParams string `json:"variation_params"`
	Status          string `json:"status,omitempty"`
	SummaryJson     string `json:"summary_json,omitempty"`
	RunDurationMs   int64  `json:"run_duration_ms,omitempty"`
}

// BestConfigurationResponse represents the best configuration found
type BestConfigurationResponse struct {
	SweepID         int64              `json:"sweep_id"`
	SimulationID    int64              `json:"simulation_id"`
	Recipe          RecipeResponse     `json:"recipe"`
	Metrics         map[string]float64 `json:"metrics"`
	Rank            int                `json:"rank"`
	Percentile      float64            `json:"percentile"`
	VariationParams map[string]any     `json:"variation_params"`
}

// RecipeResponse represents a recipe for API responses
type RecipeResponse struct {
	Version    string `json:"version"`
	Name       string `json:"name"`
	Parameters any    `json:"parameters"`
}

// VisualizationDataResponse represents visualization data for API responses
type VisualizationDataResponse struct {
	SweepID    int64                            `json:"sweep_id"`
	Parameters []string                         `json:"parameters"`
	Metrics    []string                         `json:"metrics"`
	DataPoints []VisualizationDataPointResponse `json:"data_points"`
}

// VisualizationDataPointResponse represents a data point for visualization
type VisualizationDataPointResponse struct {
	Params  map[string]any     `json:"params"`
	Metrics map[string]float64 `json:"metrics"`
}

// LeaderboardEntryResponse represents a leaderboard entry
type LeaderboardEntryResponse struct {
	Rank         int     `json:"rank"`
	SimulationID int64   `json:"simulation_id"`
	RecipeName   string  `json:"recipe_name"`
	MetricValue  float64 `json:"metric_value"`
	CreatedAt    string  `json:"created_at"`
}
