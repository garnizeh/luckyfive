package sweep

import (
	"fmt"
	"maps"
	"math"
	"sort"
)

// Generator handles the generation of parameter combinations for sweeps.
type Generator struct{}

// NewGenerator creates a new sweep generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GeneratedRecipe represents a single generated recipe from a sweep.
type GeneratedRecipe struct {
	ID          string
	Name        string
	Parameters  map[string]any
	ParentSweep string
}

// ToServiceRecipe converts a GeneratedRecipe to a service-compatible Recipe.
func (gr *GeneratedRecipe) ToServiceRecipe() Recipe {
	return Recipe{
		Version:    "1.0",
		Name:       gr.Name,
		Parameters: gr.Parameters,
	}
}

// Generate creates all valid parameter combinations for a sweep configuration.
func (g *Generator) Generate(cfg SweepConfig) ([]GeneratedRecipe, error) {
	// Validate the configuration first
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sweep config: %w", err)
	}

	// Parse parameter ranges into value sets
	paramSets := make(map[string][]float64)
	for _, param := range cfg.Parameters {
		values, err := g.expandParameter(param)
		if err != nil {
			return nil, fmt.Errorf("expand parameter %s: %w", param.Name, err)
		}
		paramSets[param.Name] = values
	}

	// Generate cartesian product
	combinations := g.cartesianProduct(paramSets)

	// Apply constraints if any
	if len(cfg.Constraints) > 0 {
		combinations = g.applyConstraints(combinations, cfg.Constraints)
	}

	// Create recipes from combinations
	recipes := make([]GeneratedRecipe, 0, len(combinations))
	for i, combo := range combinations {
		recipe := g.buildRecipe(cfg.BaseRecipe, combo, i)
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

// expandParameter converts a ParameterSweep into a slice of float64 values.
func (g *Generator) expandParameter(param ParameterSweep) ([]float64, error) {
	switch param.Type {
	case "range":
		return g.expandRange(param.Values)
	case "discrete":
		return g.expandDiscrete(param.Values)
	case "exponential":
		return g.expandExponential(param.Values)
	default:
		return nil, fmt.Errorf("unknown parameter type: %s", param.Type)
	}
}

// expandRange expands a RangeValues into a slice of values.
func (g *Generator) expandRange(values any) ([]float64, error) {
	rv, ok := values.(RangeValues)
	if !ok {
		return nil, fmt.Errorf("values must be RangeValues for range type")
	}

	var result []float64
	for v := rv.Min; v <= rv.Max; v += rv.Step {
		result = append(result, v)
	}
	return result, nil
}

// expandDiscrete expands a DiscreteValues into a slice of values.
func (g *Generator) expandDiscrete(values any) ([]float64, error) {
	dv, ok := values.(DiscreteValues)
	if !ok {
		return nil, fmt.Errorf("values must be DiscreteValues for discrete type")
	}
	return dv.Values, nil
}

// expandExponential expands an ExponentialValues into a slice of values.
func (g *Generator) expandExponential(values any) ([]float64, error) {
	ev, ok := values.(ExponentialValues)
	if !ok {
		return nil, fmt.Errorf("values must be ExponentialValues for exponential type")
	}

	var result []float64
	for i := ev.Start; i <= ev.End; i++ {
		value := math.Pow(ev.Base, float64(i))
		result = append(result, value)
	}
	return result, nil
}

// cartesianProduct generates all combinations of parameter values.
func (g *Generator) cartesianProduct(paramSets map[string][]float64) []map[string]float64 {
	// Get parameter names in stable order for consistent results
	params := make([]string, 0, len(paramSets))
	for name := range paramSets {
		params = append(params, name)
	}
	sort.Strings(params)

	// Build combinations recursively
	return g.cartesianRecursive(params, paramSets, 0, map[string]float64{})
}

// cartesianRecursive recursively builds cartesian product combinations.
func (g *Generator) cartesianRecursive(
	params []string,
	paramSets map[string][]float64,
	index int,
	current map[string]float64,
) []map[string]float64 {
	if index == len(params) {
		// Copy current combination
		combo := make(map[string]float64)
		maps.Copy(combo, current)
		return []map[string]float64{combo}
	}

	param := params[index]
	values := paramSets[param]

	var results []map[string]float64
	for _, value := range values {
		current[param] = value
		subResults := g.cartesianRecursive(params, paramSets, index+1, current)
		results = append(results, subResults...)
	}

	return results
}

// applyConstraints filters combinations based on constraints.
func (g *Generator) applyConstraints(
	combinations []map[string]float64,
	constraints []Constraint,
) []map[string]float64 {
	var filtered []map[string]float64

	for _, combo := range combinations {
		valid := true
		for _, constraint := range constraints {
			if !g.checkConstraint(combo, constraint) {
				valid = false
				break
			}
		}
		if valid {
			filtered = append(filtered, combo)
		}
	}

	return filtered
}

// checkConstraint validates a single constraint against a parameter combination.
func (g *Generator) checkConstraint(combo map[string]float64, c Constraint) bool {
	switch c.Type {
	case "sum":
		sum := 0.0
		for _, param := range c.Parameters {
			sum += combo[param]
		}
		return math.Abs(sum-c.Value) < 0.001

	case "ratio":
		if len(c.Parameters) != 2 {
			return false
		}
		denom := combo[c.Parameters[1]]
		if denom == 0 {
			return false // Avoid division by zero
		}
		ratio := combo[c.Parameters[0]] / denom
		return math.Abs(ratio-c.Value) < 0.001

	case "min":
		for _, param := range c.Parameters {
			if combo[param] < c.Value {
				return false
			}
		}
		return true

	case "max":
		for _, param := range c.Parameters {
			if combo[param] > c.Value {
				return false
			}
		}
		return true

	default:
		return true // Unknown constraints are ignored (pass through)
	}
}

// buildRecipe creates a GeneratedRecipe from base recipe and parameter combination.
func (g *Generator) buildRecipe(baseRecipe Recipe, combo map[string]float64, index int) GeneratedRecipe {
	// Copy base parameters
	params := make(map[string]any)
	for k, v := range baseRecipe.Parameters {
		params[k] = v
	}

	// Override with sweep values
	for paramName, value := range combo {
		params[paramName] = value
	}

	return GeneratedRecipe{
		ID:          fmt.Sprintf("sweep_var_%d", index),
		Name:        fmt.Sprintf("%s_var_%d", baseRecipe.Name, index),
		Parameters:  params,
		ParentSweep: "", // Will be set by caller
	}
}
