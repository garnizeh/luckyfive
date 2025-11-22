package sweep

import (
	"testing"
)

func TestGenerator_Generate_SimpleRange(t *testing.T) {
	gen := NewGenerator()

	cfg := SweepConfig{
		Name: "test_sweep",
		BaseRecipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: map[string]any{
				"alpha": 0.5,
				"beta":  0.3,
			},
		},
		Parameters: []ParameterSweep{
			{
				Name: "alpha",
				Type: "range",
				Values: RangeValues{
					Min:  0.0,
					Max:  1.0,
					Step: 0.5,
				},
			},
		},
	}

	recipes, err := gen.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expected := 3 // 0.0, 0.5, 1.0
	if len(recipes) != expected {
		t.Errorf("Expected %d recipes, got %d", expected, len(recipes))
	}

	// Check that alpha values are correct
	expectedAlphas := []float64{0.0, 0.5, 1.0}
	for i, recipe := range recipes {
		alpha, ok := recipe.Parameters["alpha"].(float64)
		if !ok {
			t.Errorf("Recipe %d: alpha not found or not float64", i)
			continue
		}
		if alpha != expectedAlphas[i] {
			t.Errorf("Recipe %d: expected alpha %f, got %f", i, expectedAlphas[i], alpha)
		}
		// Beta should remain from base recipe
		if beta, ok := recipe.Parameters["beta"].(float64); !ok || beta != 0.3 {
			t.Errorf("Recipe %d: expected beta 0.3, got %v", i, beta)
		}
	}
}

func TestGenerator_Generate_Discrete(t *testing.T) {
	gen := NewGenerator()

	cfg := SweepConfig{
		Name: "discrete_test",
		BaseRecipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: map[string]any{
				"count": 100,
			},
		},
		Parameters: []ParameterSweep{
			{
				Name: "count",
				Type: "discrete",
				Values: DiscreteValues{
					Values: []float64{10, 50, 100, 500},
				},
			},
		},
	}

	recipes, err := gen.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(recipes) != 4 {
		t.Errorf("Expected 4 recipes, got %d", len(recipes))
	}

	expectedCounts := []float64{10, 50, 100, 500}
	for i, recipe := range recipes {
		count, ok := recipe.Parameters["count"].(float64)
		if !ok || count != expectedCounts[i] {
			t.Errorf("Recipe %d: expected count %f, got %v", i, expectedCounts[i], count)
		}
	}
}

func TestGenerator_Generate_Exponential(t *testing.T) {
	gen := NewGenerator()

	cfg := SweepConfig{
		Name: "exp_test",
		BaseRecipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: map[string]any{
				"size": 1000,
			},
		},
		Parameters: []ParameterSweep{
			{
				Name: "size",
				Type: "exponential",
				Values: ExponentialValues{
					Base:  10,
					Start: 1,
					End:   3,
				},
			},
		},
	}

	recipes, err := gen.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(recipes) != 3 {
		t.Errorf("Expected 3 recipes, got %d", len(recipes))
	}

	expectedSizes := []float64{10, 100, 1000}
	for i, recipe := range recipes {
		size, ok := recipe.Parameters["size"].(float64)
		if !ok || size != expectedSizes[i] {
			t.Errorf("Recipe %d: expected size %f, got %v", i, expectedSizes[i], size)
		}
	}
}

func TestGenerator_Generate_CartesianProduct(t *testing.T) {
	gen := NewGenerator()

	cfg := SweepConfig{
		Name: "cartesian_test",
		BaseRecipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: map[string]any{
				"fixed": 42,
			},
		},
		Parameters: []ParameterSweep{
			{
				Name:   "alpha",
				Type:   "range",
				Values: RangeValues{Min: 0.0, Max: 1.0, Step: 1.0},
			},
			{
				Name:   "beta",
				Type:   "discrete",
				Values: DiscreteValues{Values: []float64{0.1, 0.2}},
			},
		},
	}

	recipes, err := gen.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should be 2 * 2 = 4 combinations
	if len(recipes) != 4 {
		t.Errorf("Expected 4 recipes, got %d", len(recipes))
	}

	// Check that all combinations are present
	expectedCombos := []struct{ alpha, beta float64 }{
		{0.0, 0.1}, {0.0, 0.2}, {1.0, 0.1}, {1.0, 0.2},
	}

	for _, expected := range expectedCombos {
		found := false
		for _, recipe := range recipes {
			alpha := recipe.Parameters["alpha"].(float64)
			beta := recipe.Parameters["beta"].(float64)
			if alpha == expected.alpha && beta == expected.beta {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected combination alpha=%f, beta=%f not found", expected.alpha, expected.beta)
		}
	}
}

func TestGenerator_Generate_WithConstraints(t *testing.T) {
	gen := NewGenerator()

	cfg := SweepConfig{
		Name: "constraint_test",
		BaseRecipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: map[string]any{
				"gamma": 0.2,
			},
		},
		Parameters: []ParameterSweep{
			{
				Name:   "alpha",
				Type:   "range",
				Values: RangeValues{Min: 0.0, Max: 1.0, Step: 0.5},
			},
			{
				Name:   "beta",
				Type:   "range",
				Values: RangeValues{Min: 0.0, Max: 1.0, Step: 0.5},
			},
		},
		Constraints: []Constraint{
			{
				Type:       "sum",
				Parameters: []string{"alpha", "beta"},
				Value:      1.0,
			},
		},
	}

	recipes, err := gen.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Without constraints: 3 * 3 = 9 combinations
	// With sum constraint: only combinations where alpha + beta = 1.0
	expected := 3 // (0.0,1.0), (0.5,0.5), (1.0,0.0)
	if len(recipes) != expected {
		t.Errorf("Expected %d recipes with constraints, got %d", expected, len(recipes))
	}

	// Verify all combinations sum to 1.0
	for i, recipe := range recipes {
		alpha := recipe.Parameters["alpha"].(float64)
		beta := recipe.Parameters["beta"].(float64)
		sum := alpha + beta
		if sum < 0.999 || sum > 1.001 { // Allow small floating point error
			t.Errorf("Recipe %d: alpha + beta = %f, expected 1.0", i, sum)
		}
	}
}

func TestGenerator_Generate_InvalidConfig(t *testing.T) {
	gen := NewGenerator()

	// Test invalid config (empty name)
	cfg := SweepConfig{
		Name: "", // Invalid
		BaseRecipe: Recipe{
			Version:    "1.0",
			Name:       "test",
			Parameters: map[string]any{"test": 1},
		},
		Parameters: []ParameterSweep{
			{
				Name:   "test",
				Type:   "range",
				Values: RangeValues{Min: 0, Max: 1, Step: 0.1},
			},
		},
	}

	_, err := gen.Generate(cfg)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestExpandParameter_InvalidType(t *testing.T) {
	gen := NewGenerator()

	param := ParameterSweep{
		Name:   "test",
		Type:   "invalid",
		Values: "bad",
	}

	_, err := gen.expandParameter(param)
	if err == nil {
		t.Error("Expected error for invalid parameter type")
	}
}

func BenchmarkGenerator_GenerateLargeSweep(b *testing.B) {
	gen := NewGenerator()

	cfg := SweepConfig{
		Name: "large_sweep",
		BaseRecipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: map[string]any{
				"fixed": 1,
			},
		},
		Parameters: []ParameterSweep{
			{
				Name:   "p1",
				Type:   "range",
				Values: RangeValues{Min: 0, Max: 10, Step: 1},
			},
			{
				Name:   "p2",
				Type:   "range",
				Values: RangeValues{Min: 0, Max: 10, Step: 1},
			},
			{
				Name:   "p3",
				Type:   "range",
				Values: RangeValues{Min: 0, Max: 10, Step: 1},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(cfg)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}
