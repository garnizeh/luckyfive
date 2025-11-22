package sweep

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SweepConfig defines the configuration for a parameter sweep
type SweepConfig struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	BaseRecipe  Recipe           `json:"base_recipe"`
	Parameters  []ParameterSweep `json:"parameters"`
	Constraints []Constraint     `json:"constraints,omitempty"`
}

// ParameterSweep defines how to sweep a single parameter
type ParameterSweep struct {
	Name   string `json:"name"` // e.g., "alpha", "sim_prev_max"
	Type   string `json:"type"` // "range", "discrete", "exponential"
	Values any    `json:"values"`
}

// MarshalJSON implements custom JSON marshaling for ParameterSweep
func (p ParameterSweep) MarshalJSON() ([]byte, error) {
	type Alias ParameterSweep
	aux := &struct {
		Alias
		Values json.RawMessage `json:"values"`
	}{
		Alias: Alias(p),
	}

	switch p.Type {
	case "range":
		if rv, ok := p.Values.(RangeValues); ok {
			values, err := json.Marshal(rv)
			if err != nil {
				return nil, err
			}
			aux.Values = values
		}
	case "discrete":
		if dv, ok := p.Values.(DiscreteValues); ok {
			values, err := json.Marshal(dv)
			if err != nil {
				return nil, err
			}
			aux.Values = values
		}
	case "exponential":
		if ev, ok := p.Values.(ExponentialValues); ok {
			values, err := json.Marshal(ev)
			if err != nil {
				return nil, err
			}
			aux.Values = values
		}
	}

	return json.Marshal(aux)
}

// UnmarshalJSON implements custom JSON unmarshaling for ParameterSweep
func (p *ParameterSweep) UnmarshalJSON(data []byte) error {
	type Alias ParameterSweep
	aux := &struct {
		Alias
		Values json.RawMessage `json:"values"`
	}{}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	p.Name = aux.Name
	p.Type = aux.Type

	switch p.Type {
	case "range":
		var rv RangeValues
		if err := json.Unmarshal(aux.Values, &rv); err != nil {
			return err
		}
		p.Values = rv
	case "discrete":
		var dv DiscreteValues
		if err := json.Unmarshal(aux.Values, &dv); err != nil {
			return err
		}
		p.Values = dv
	case "exponential":
		var ev ExponentialValues
		if err := json.Unmarshal(aux.Values, &ev); err != nil {
			return err
		}
		p.Values = ev
	default:
		return fmt.Errorf("unknown parameter type: %s", p.Type)
	}

	return nil
}

// RangeValues defines a range of values for a parameter
type RangeValues struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Step float64 `json:"step"`
}

// DiscreteValues defines a discrete set of values for a parameter
type DiscreteValues struct {
	Values []float64 `json:"values"`
}

// ExponentialValues defines an exponential range of values
type ExponentialValues struct {
	Base  float64 `json:"base"`  // e.g., 10
	Start int     `json:"start"` // e.g., -3
	End   int     `json:"end"`   // e.g., 2
}

// Constraint defines a constraint on parameter combinations
type Constraint struct {
	Type       string   `json:"type"` // "sum", "ratio", "min", "max"
	Parameters []string `json:"parameters"`
	Value      float64  `json:"value"`
}

// Recipe defines a simulation recipe
type Recipe struct {
	Version    string         `json:"version"`
	Name       string         `json:"name"`
	Parameters map[string]any `json:"parameters"`
}

// Validate validates the SweepConfig
func (s *SweepConfig) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}

	if len(s.Parameters) == 0 {
		return errors.New("at least one parameter is required")
	}

	// Validate base recipe
	if err := s.BaseRecipe.Validate(); err != nil {
		return fmt.Errorf("base_recipe: %w", err)
	}

	// Validate parameters
	for i, param := range s.Parameters {
		if err := param.Validate(); err != nil {
			return fmt.Errorf("parameter %d (%s): %w", i, param.Name, err)
		}
	}

	// Validate constraints
	for i, constraint := range s.Constraints {
		if err := constraint.Validate(); err != nil {
			return fmt.Errorf("constraint %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the ParameterSweep
func (p *ParameterSweep) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}

	switch p.Type {
	case "range":
		return p.validateRangeValues()
	case "discrete":
		return p.validateDiscreteValues()
	case "exponential":
		return p.validateExponentialValues()
	default:
		return fmt.Errorf("unknown type: %s", p.Type)
	}
}

func (p *ParameterSweep) validateRangeValues() error {
	rv, ok := p.Values.(RangeValues)
	if !ok {
		return errors.New("values must be RangeValues for type 'range'")
	}

	if rv.Min >= rv.Max {
		return errors.New("min must be less than max")
	}

	if rv.Step <= 0 {
		return errors.New("step must be positive")
	}

	return nil
}

func (p *ParameterSweep) validateDiscreteValues() error {
	dv, ok := p.Values.(DiscreteValues)
	if !ok {
		return errors.New("values must be DiscreteValues for type 'discrete'")
	}

	if len(dv.Values) == 0 {
		return errors.New("at least one value is required")
	}

	return nil
}

func (p *ParameterSweep) validateExponentialValues() error {
	ev, ok := p.Values.(ExponentialValues)
	if !ok {
		return errors.New("values must be ExponentialValues for type 'exponential'")
	}

	if ev.Base <= 0 {
		return errors.New("base must be positive")
	}

	if ev.Start >= ev.End {
		return errors.New("start must be less than end")
	}

	return nil
}

// Validate validates the Constraint
func (c *Constraint) Validate() error {
	if len(c.Parameters) == 0 {
		return errors.New("at least one parameter is required")
	}

	switch c.Type {
	case "sum", "ratio", "min", "max":
		// Valid types
	default:
		return fmt.Errorf("unknown type: %s", c.Type)
	}

	return nil
}

// Validate validates the Recipe
func (r *Recipe) Validate() error {
	if r.Version == "" {
		return errors.New("version is required")
	}

	if r.Name == "" {
		return errors.New("name is required")
	}

	if len(r.Parameters) == 0 {
		return errors.New("parameters cannot be empty")
	}

	return nil
}
