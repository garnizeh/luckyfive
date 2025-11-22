# Sweep Configuration Examples

This directory contains example sweep configurations for parameter optimization in LuckyFive simulations.

## Configuration Schema

A sweep configuration defines how to vary simulation parameters systematically to find optimal settings.

### Basic Structure

```json
{
  "name": "sweep_name",
  "description": "Description of what this sweep does",
  "base_recipe": {
    "version": "1.0",
    "name": "recipe_name",
    "parameters": {
      "param1": "value1",
      "param2": "value2"
    }
  },
  "parameters": [
    {
      "name": "parameter_name",
      "type": "range|discrete|exponential",
      "values": { ... }
    }
  ],
  "constraints": [
    {
      "type": "sum|ratio|min|max",
      "parameters": ["param1", "param2"],
      "value": 1.0
    }
  ]
}
```

## Parameter Types

### Range
Sweeps a parameter over a continuous range with fixed step size.

```json
{
  "name": "alpha",
  "type": "range",
  "values": {
    "min": 0.0,
    "max": 1.0,
    "step": 0.1
  }
}
```

### Discrete
Sweeps a parameter over a specific set of values.

```json
{
  "name": "sim_prev_max",
  "type": "discrete",
  "values": {
    "values": [50, 100, 200, 500, 1000]
  }
}
```

### Exponential
Sweeps a parameter over exponential values (useful for parameters that vary by orders of magnitude).

```json
{
  "name": "alpha",
  "type": "exponential",
  "values": {
    "base": 10,
    "start": -3,
    "end": 0
  }
}
```

This generates values: 0.001, 0.01, 0.1, 1.0

## Constraints

Constraints limit parameter combinations during the sweep.

### Types
- `sum`: Parameters must sum to the specified value
- `ratio`: Ratio between parameters must equal the value
- `min`: Minimum value constraint
- `max`: Maximum value constraint

## Examples

### alpha_range.json
Simple sweep of the alpha parameter from 0.0 to 1.0 in steps of 0.1.

### multi_param.json
Multi-parameter sweep with alpha, sim_prev_max, and sim_count, including a maximum constraint on alpha.

### exponential.json
Exponential sweep for parameters that benefit from logarithmic scaling.

## Usage

These configurations can be loaded and validated using the sweep package:

```go
import "github.com/garnizeh/luckyfive/pkg/sweep"

// Load and validate a sweep config
config, err := loadSweepConfig("alpha_range.json")
if err != nil {
    log.Fatal(err)
}

if err := config.Validate(); err != nil {
    log.Fatal(err)
}
```

## Validation

All sweep configurations are validated to ensure:
- Required fields are present
- Parameter types are valid
- Value ranges are sensible
- Constraints are properly defined
- Base recipe is valid