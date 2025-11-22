package sweep

import (
	"encoding/json"
	"testing"
)

func TestSweepConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SweepConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SweepConfig{
				Name:        "test_sweep",
				Description: "Test sweep configuration",
				BaseRecipe: Recipe{
					Version: "1.0",
					Name:    "advanced",
					Parameters: map[string]any{
						"alpha": 0.1,
					},
				},
				Parameters: []ParameterSweep{
					{
						Name: "alpha",
						Type: "range",
						Values: RangeValues{
							Min:  0.0,
							Max:  1.0,
							Step: 0.1,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: SweepConfig{
				BaseRecipe: Recipe{
					Version: "1.0",
					Name:    "advanced",
					Parameters: map[string]any{
						"alpha": 0.1,
					},
				},
				Parameters: []ParameterSweep{
					{
						Name: "alpha",
						Type: "range",
						Values: RangeValues{
							Min:  0.0,
							Max:  1.0,
							Step: 0.1,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no parameters",
			config: SweepConfig{
				Name: "test_sweep",
				BaseRecipe: Recipe{
					Version: "1.0",
					Name:    "advanced",
					Parameters: map[string]any{
						"alpha": 0.1,
					},
				},
				Parameters: []ParameterSweep{},
			},
			wantErr: true,
		},
		{
			name: "invalid base recipe",
			config: SweepConfig{
				Name: "test_sweep",
				BaseRecipe: Recipe{
					Name: "advanced",
					Parameters: map[string]any{
						"alpha": 0.1,
					},
				},
				Parameters: []ParameterSweep{
					{
						Name: "alpha",
						Type: "range",
						Values: RangeValues{
							Min:  0.0,
							Max:  1.0,
							Step: 0.1,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SweepConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParameterSweep_Validate(t *testing.T) {
	tests := []struct {
		name    string
		param   ParameterSweep
		wantErr bool
	}{
		{
			name: "valid range",
			param: ParameterSweep{
				Name: "alpha",
				Type: "range",
				Values: RangeValues{
					Min:  0.0,
					Max:  1.0,
					Step: 0.1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid discrete",
			param: ParameterSweep{
				Name: "beta",
				Type: "discrete",
				Values: DiscreteValues{
					Values: []float64{0.1, 0.5, 0.9},
				},
			},
			wantErr: false,
		},
		{
			name: "valid exponential",
			param: ParameterSweep{
				Name: "gamma",
				Type: "exponential",
				Values: ExponentialValues{
					Base:  10,
					Start: -2,
					End:   2,
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			param: ParameterSweep{
				Type: "range",
				Values: RangeValues{
					Min:  0.0,
					Max:  1.0,
					Step: 0.1,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			param: ParameterSweep{
				Name: "alpha",
				Type: "invalid",
				Values: RangeValues{
					Min:  0.0,
					Max:  1.0,
					Step: 0.1,
				},
			},
			wantErr: true,
		},
		{
			name: "range min >= max",
			param: ParameterSweep{
				Name: "alpha",
				Type: "range",
				Values: RangeValues{
					Min:  1.0,
					Max:  1.0,
					Step: 0.1,
				},
			},
			wantErr: true,
		},
		{
			name: "range step <= 0",
			param: ParameterSweep{
				Name: "alpha",
				Type: "range",
				Values: RangeValues{
					Min:  0.0,
					Max:  1.0,
					Step: 0.0,
				},
			},
			wantErr: true,
		},
		{
			name: "discrete empty values",
			param: ParameterSweep{
				Name: "beta",
				Type: "discrete",
				Values: DiscreteValues{
					Values: []float64{},
				},
			},
			wantErr: true,
		},
		{
			name: "exponential base <= 0",
			param: ParameterSweep{
				Name: "gamma",
				Type: "exponential",
				Values: ExponentialValues{
					Base:  0,
					Start: -2,
					End:   2,
				},
			},
			wantErr: true,
		},
		{
			name: "exponential start >= end",
			param: ParameterSweep{
				Name: "gamma",
				Type: "exponential",
				Values: ExponentialValues{
					Base:  10,
					Start: 2,
					End:   2,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.param.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParameterSweep.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecipe_Validate(t *testing.T) {
	tests := []struct {
		name    string
		recipe  Recipe
		wantErr bool
	}{
		{
			name: "valid recipe",
			recipe: Recipe{
				Version: "1.0",
				Name:    "advanced",
				Parameters: map[string]any{
					"alpha": 0.1,
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			recipe: Recipe{
				Name: "advanced",
				Parameters: map[string]any{
					"alpha": 0.1,
				},
			},
			wantErr: true,
		},
		{
			name: "missing name",
			recipe: Recipe{
				Version: "1.0",
				Parameters: map[string]any{
					"alpha": 0.1,
				},
			},
			wantErr: true,
		},
		{
			name: "empty parameters",
			recipe: Recipe{
				Version:    "1.0",
				Name:       "advanced",
				Parameters: map[string]any{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.recipe.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Recipe.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConstraint_Validate(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
		wantErr    bool
	}{
		{
			name: "valid sum constraint",
			constraint: Constraint{
				Type:       "sum",
				Parameters: []string{"alpha", "beta"},
				Value:      1.0,
			},
			wantErr: false,
		},
		{
			name: "valid ratio constraint",
			constraint: Constraint{
				Type:       "ratio",
				Parameters: []string{"alpha", "beta"},
				Value:      2.0,
			},
			wantErr: false,
		},
		{
			name: "no parameters",
			constraint: Constraint{
				Type:  "sum",
				Value: 1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			constraint: Constraint{
				Type:       "invalid",
				Parameters: []string{"alpha"},
				Value:      1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constraint.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Constraint.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONMarshaling(t *testing.T) {
	config := SweepConfig{
		Name:        "alpha_sweep",
		Description: "Sweep alpha parameter from 0.0 to 1.0",
		BaseRecipe: Recipe{
			Version: "1.0",
			Name:    "advanced",
			Parameters: map[string]any{
				"alpha":        0.1,
				"sim_prev_max": 100,
			},
		},
		Parameters: []ParameterSweep{
			{
				Name: "alpha",
				Type: "range",
				Values: RangeValues{
					Min:  0.0,
					Max:  1.0,
					Step: 0.1,
				},
			},
			{
				Name: "sim_prev_max",
				Type: "discrete",
				Values: DiscreteValues{
					Values: []float64{50, 100, 200, 500},
				},
			},
		},
		Constraints: []Constraint{
			{
				Type:       "max",
				Parameters: []string{"alpha"},
				Value:      0.8,
			},
		},
	}

	// Test marshaling
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Test unmarshaling
	var unmarshaled SweepConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Validate unmarshaled config
	if err := unmarshaled.Validate(); err != nil {
		t.Fatalf("Unmarshaled config is invalid: %v", err)
	}

	// Check key fields
	if unmarshaled.Name != config.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, config.Name)
	}

	if len(unmarshaled.Parameters) != len(config.Parameters) {
		t.Errorf("Parameters length mismatch: got %d, want %d", len(unmarshaled.Parameters), len(config.Parameters))
	}
}
