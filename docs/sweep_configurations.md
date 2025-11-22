# Sweep Configurations Guide

## Overview

Sweep configurations enable systematic parameter optimization for lottery simulations. A sweep defines how to vary simulation parameters across multiple runs to find optimal configurations. This guide explains how to create and configure sweep configurations for the LuckyFive system.

## What is a Parameter Sweep?

A parameter sweep systematically explores different combinations of simulation parameters to:
- Find optimal parameter values for better prediction accuracy
- Understand parameter sensitivity and interactions
- Compare different algorithmic approaches
- Validate model robustness across parameter ranges

## Configuration Structure

### Basic Structure

```json
{
  "name": "my_sweep_config",
  "description": "Description of what this sweep optimizes",
  "base_recipe": {
    "version": "1.0",
    "name": "advanced",
    "parameters": {
      "alpha": 0.1,
      "sim_prev_max": 100,
      "sim_count": 1000,
      "scorer_type": "frequency"
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

## Base Recipe Parameters

The `base_recipe` defines the simulation algorithm and default parameter values. All sweep variations start from this baseline.

### Core Parameters

#### `alpha` (Float: 0.0 - 1.0)
**Purpose:** Controls the weight given to frequency-based scoring vs recency-based scoring.
- `0.0`: Pure frequency scoring (how often numbers appear)
- `1.0`: Pure recency scoring (how recently numbers appeared)
- `0.5`: Balanced approach

**Impact:** Higher alpha favors recently drawn numbers, lower alpha favors historically frequent numbers.

#### `sim_prev_max` (Integer: 1 - 1000+)
**Purpose:** Maximum number of previous contests to analyze for patterns.
- Small values (10-50): Focus on very recent patterns
- Large values (200-500): Consider longer historical trends
- Very large values (1000+): Include entire available history

**Impact:** Larger values provide more data but may include outdated patterns.

#### `sim_count` (Integer: 100 - 10000+)
**Purpose:** Number of simulation runs to perform for each parameter combination.
- Small values (100-500): Fast results, higher variance
- Large values (1000-5000): More stable results, slower execution
- Very large values (10000+): High precision, very slow

**Impact:** More simulations reduce random variation but increase computation time.

#### `scorer_type` (String: "frequency", "evolutionary", "advanced")
**Purpose:** Algorithm used to score lottery numbers.
- `"frequency"`: Simple frequency analysis
- `"evolutionary"`: Adaptive algorithm that evolves over time
- `"advanced"`: Multi-factor scoring with advanced heuristics

**Impact:** Different scorers have different strengths for different lottery patterns.

### Advanced Parameters

#### `beta` (Float: 0.0 - 2.0)
**Purpose:** Controls the steepness of scoring curves in advanced scorer.
- Low values (< 0.5): Gentle scoring transitions
- Medium values (0.5-1.0): Standard scoring
- High values (> 1.0): Sharp scoring transitions

#### `lambda` (Float: 0.0 - 1.0)
**Purpose:** Decay factor for recency weighting.
- `0.0`: No decay (all historical data equally weighted)
- `0.5`: Moderate decay (recent data more important)
- `1.0`: Strong decay (only very recent data matters)

#### `gamma` (Float: 0.0 - 2.0)
**Purpose:** Clustering sensitivity in advanced scorer.
- Low values: Less sensitive to number clustering
- High values: More sensitive to number patterns

#### `cluster_threshold` (Float: 0.0 - 1.0)
**Purpose:** Minimum similarity score for number clustering.
- Low values: More clusters formed
- High values: Fewer, tighter clusters

#### `hot_cold_weight` (Float: 0.0 - 1.0)
**Purpose:** Weight given to hot/cold number analysis.
- `0.0`: Ignore hot/cold patterns
- `1.0`: Strong hot/cold weighting

#### `cooccurrence_weight` (Float: 0.0 - 1.0)
**Purpose:** Weight given to number co-occurrence patterns.
- `0.0`: Ignore number relationships
- `1.0`: Strong co-occurrence weighting

## Parameter Sweep Types

### Range Parameters

Sweep a parameter continuously over a range with fixed step size.

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

**Use when:** Parameter has linear impact, you want fine-grained control.

### Discrete Parameters

Sweep over specific predefined values.

```json
{
  "name": "sim_prev_max",
  "type": "discrete",
  "values": {
    "values": [50, 100, 200, 500, 1000]
  }
}
```

**Use when:** Only specific values make sense, or you want to test known good values.

### Exponential Parameters

Sweep over exponentially increasing values (useful for parameters that vary by orders of magnitude).

```json
{
  "name": "sim_count",
  "type": "exponential",
  "values": {
    "base": 10,
    "start": 2,
    "end": 4
  }
}
```

This generates: 100, 1000, 10000

**Use when:** Parameter impact scales logarithmically.

## Constraints

Constraints limit parameter combinations to prevent invalid or nonsensical configurations.

### Sum Constraint
Parameters must sum to a specific value.

```json
{
  "type": "sum",
  "parameters": ["alpha", "beta", "gamma"],
  "value": 1.0
}
```

### Ratio Constraint
Ratio between parameters must equal a value.

```json
{
  "type": "ratio",
  "parameters": ["alpha", "beta"],
  "value": 2.0
}
```

### Min/Max Constraints
Parameter must be above/below a threshold.

```json
{
  "type": "max",
  "parameters": ["alpha"],
  "value": 0.8
}
```

## Example Configurations

### Simple Alpha Sweep
```json
{
  "name": "alpha_optimization",
  "description": "Find optimal alpha value for frequency vs recency balance",
  "base_recipe": {
    "version": "1.0",
    "name": "advanced",
    "parameters": {
      "alpha": 0.1,
      "sim_prev_max": 100,
      "sim_count": 1000,
      "scorer_type": "frequency"
    }
  },
  "parameters": [
    {
      "name": "alpha",
      "type": "range",
      "values": {
        "min": 0.0,
        "max": 1.0,
        "step": 0.1
      }
    }
  ]
}
```

### Multi-Parameter Optimization
```json
{
  "name": "comprehensive_optimization",
  "description": "Optimize multiple parameters simultaneously",
  "base_recipe": {
    "version": "1.0",
    "name": "advanced",
    "parameters": {
      "alpha": 0.1,
      "sim_prev_max": 100,
      "sim_count": 1000,
      "scorer_type": "advanced"
    }
  },
  "parameters": [
    {
      "name": "alpha",
      "type": "range",
      "values": {
        "min": 0.0,
        "max": 0.8,
        "step": 0.1
      }
    },
    {
      "name": "sim_prev_max",
      "type": "discrete",
      "values": {
        "values": [50, 100, 200, 500]
      }
    },
    {
      "name": "sim_count",
      "type": "exponential",
      "values": {
        "base": 10,
        "start": 2,
        "end": 4
      }
    }
  ],
  "constraints": [
    {
      "type": "max",
      "parameters": ["alpha"],
      "value": 0.8
    }
  ]
}
```

## Best Practices

### Parameter Selection
1. **Start Simple:** Begin with single parameter sweeps to understand individual impacts
2. **Use Appropriate Ranges:** Don't sweep over nonsensical parameter values
3. **Consider Computation Cost:** Larger sweeps take exponentially more time
4. **Validate Results:** Always test top-performing configurations manually

### Common Sweep Patterns

#### Alpha Optimization
- Range: 0.0 to 1.0, step 0.1
- Purpose: Find balance between frequency and recency

#### History Window Optimization
- Values: [20, 50, 100, 200, 500, 1000]
- Purpose: Determine optimal historical analysis window

#### Simulation Count Scaling
- Exponential: base 10, start 2, end 4 (100-10000)
- Purpose: Balance accuracy vs computation time

### Performance Considerations

- **Cartesian Product Growth:** Each additional parameter multiplies total combinations
- **Simulation Count Impact:** Higher sim_count dramatically increases runtime
- **Parallel Execution:** Sweeps can be executed in parallel for better performance

## API Usage

### Creating a Sweep Configuration
```bash
POST /api/v1/sweep-configs
Content-Type: application/json

{
  "name": "my_sweep",
  "description": "My optimization sweep",
  "config": { ... sweep configuration ... }
}
```

### Listing Configurations
```bash
GET /api/v1/sweep-configs
```

### Getting a Specific Configuration
```bash
GET /api/v1/sweep-configs/{id}
```

## Validation

All sweep configurations are automatically validated to ensure:
- Required fields are present
- Parameter types are valid
- Value ranges are sensible
- Constraints are properly defined
- Base recipe parameters are valid

Invalid configurations will be rejected with detailed error messages.</content>
<parameter name="filePath">g:\code\Go\copilot\luckyfive\docs\sweep_configurations.md