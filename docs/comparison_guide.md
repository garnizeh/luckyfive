# Comparison Engine Guide

## Overview

The Comparison Engine allows you to compare multiple lottery simulations across various performance metrics to identify which configurations perform best. This guide explains how to use the comparison features to analyze and optimize your lottery prediction strategies.

## What is Simulation Comparison?

Simulation comparison enables:
- **Performance Analysis:** Compare different parameter configurations side-by-side
- **Strategy Validation:** Test which approaches work better for specific lottery patterns
- **Optimization Insights:** Identify top-performing parameter combinations
- **Statistical Analysis:** Get detailed statistics and rankings across multiple metrics

## Available Metrics

The comparison engine supports the following performance metrics:

### Hit Rate Metrics
- **`quina_rate`**: Percentage of contests where at least one quina (5 numbers) was correctly predicted
- **`quadra_rate`**: Percentage of contests where at least one quadra (4 numbers) was correctly predicted
- **`terno_rate`**: Percentage of contests where at least one terno (3 numbers) was correctly predicted

### Volume Metrics
- **`avg_hits`**: Average number of correct numbers predicted per contest
- **`total_quinaz`**: Total number of quinas hit across all contests
- **`total_quadras`**: Total number of quadras hit across all contests
- **`total_ternos`**: Total number of ternos hit across all contests

### Efficiency Metrics
- **`hit_efficiency`**: Average hits per contest (same as avg_hits for completed simulations)

## Creating Comparisons

### Basic Comparison Request

```json
{
  "name": "alpha_parameter_comparison",
  "description": "Comparing different alpha values for prediction accuracy",
  "simulation_ids": [1, 2, 3, 4, 5],
  "metrics": ["quina_rate", "avg_hits"]
}
```

### Parameters

- **`name`** (required): Human-readable name for the comparison
- **`description`** (optional): Detailed description of what is being compared
- **`simulation_ids`** (required): Array of at least 2 simulation IDs to compare
- **`metrics`** (optional): Array of metrics to compare. Defaults to `["quina_rate", "avg_hits"]`

## Comparison Results

### Result Structure

```json
{
  "id": 123,
  "name": "alpha_parameter_comparison",
  "description": "Comparing different alpha values",
  "simulation_ids": [1, 2, 3, 4, 5],
  "metrics": ["quina_rate", "avg_hits"],
  "rankings": {
    "quina_rate": [
      {
        "simulation_id": 2,
        "simulation_name": "alpha_0.3",
        "value": 0.85,
        "rank": 1,
        "percentile": 100.0
      },
      {
        "simulation_id": 4,
        "simulation_name": "alpha_0.5",
        "value": 0.82,
        "rank": 2,
        "percentile": 80.0
      }
    ],
    "avg_hits": [
      {
        "simulation_id": 2,
        "simulation_name": "alpha_0.3",
        "value": 3.2,
        "rank": 1,
        "percentile": 100.0
      }
    ]
  },
  "statistics": {
    "quina_rate": {
      "mean": 0.75,
      "median": 0.78,
      "std_dev": 0.08,
      "min": 0.65,
      "max": 0.85,
      "count": 5
    }
  },
  "winner_by_metric": {
    "quina_rate": 2,
    "avg_hits": 2
  },
  "created_at": "2025-11-22T10:30:00Z"
}
```

### Rankings

For each metric, simulations are ranked by performance (highest value = rank 1):
- **`simulation_id`**: Unique identifier of the simulation
- **`simulation_name`**: Human-readable name from the simulation recipe
- **`value`**: Raw metric value
- **`rank`**: Position in the ranking (1 = best)
- **`percentile`**: Performance percentile (100.0 = best, 0.0 = worst)

### Statistics

Statistical summary for each metric:
- **`mean`**: Average value across all simulations
- **`median`**: Middle value when sorted
- **`std_dev`**: Standard deviation (variability measure)
- **`min`**: Lowest value
- **`max`**: Highest value
- **`count`**: Number of simulations compared

### Winners

- **`winner_by_metric`**: Simulation ID that performed best for each metric

## API Usage

### Creating a Comparison

```bash
POST /api/v1/comparisons
Content-Type: application/json

{
  "name": "parameter_study",
  "description": "Comparing different parameter configurations",
  "simulation_ids": [10, 11, 12, 13],
  "metrics": ["quina_rate", "quadra_rate", "avg_hits"]
}
```

**Response:** HTTP 201 with full comparison results

### Retrieving a Comparison

```bash
GET /api/v1/comparisons/{id}
```

**Response:** HTTP 200 with comparison details

### Listing Comparisons

```bash
GET /api/v1/comparisons?limit=10&offset=0
```

**Response:** HTTP 200 with paginated list of comparisons

## Best Practices

### Metric Selection

1. **Primary Metrics:** Start with `quina_rate` and `avg_hits` for overall performance
2. **Specific Analysis:** Use `quadra_rate` and `terno_rate` for detailed hit pattern analysis
3. **Volume Metrics:** Use `total_*` metrics when comparing across different contest ranges

### Simulation Selection

1. **Completed Only:** Only completed simulations can be compared
2. **Consistent Range:** Compare simulations run over the same contest ranges for fair comparison
3. **Parameter Variation:** Group simulations by the parameter you're testing (e.g., all alpha variations)

### Statistical Interpretation

1. **Standard Deviation:** High std_dev indicates parameter sensitivity
2. **Percentiles:** Focus on top 25% performers for optimization
3. **Consistency:** Look for simulations that rank well across multiple metrics

## Example Workflows

### Parameter Optimization Study

1. Create multiple simulations with varying alpha values (0.1, 0.3, 0.5, 0.7, 0.9)
2. Run all simulations over the same contest range
3. Compare using `["quina_rate", "avg_hits", "hit_efficiency"]`
4. Identify the optimal alpha value from the rankings

### Algorithm Comparison

1. Create simulations using different scorer types ("frequency", "advanced", "evolutionary")
2. Use identical parameters except for the scorer_type
3. Compare across all available metrics
4. Determine which algorithm performs best for your lottery

### History Window Analysis

1. Create simulations with different `sim_prev_max` values (50, 100, 200, 500, 1000)
2. Compare to find the optimal historical analysis window
3. Consider both accuracy and computational cost

## Integration with Sweeps

Comparisons work seamlessly with parameter sweeps:

1. Run a parameter sweep to generate multiple simulation variations
2. Use the comparison API to analyze all sweep results
3. Identify the best-performing parameter combinations
4. Use sweep visualization endpoints to explore parameter interactions

## Error Handling

### Common Errors

- **Incomplete Simulations:** Only completed simulations can be compared
- **Invalid Metrics:** Requested metrics must be from the supported list
- **Minimum Simulations:** At least 2 simulations required for comparison
- **Not Found:** Comparison or simulation IDs must exist

### Error Responses

```json
{
  "code": "comparison_failed",
  "message": "simulation 123 is not completed (status: running)"
}
```

## Performance Considerations

- **Database Queries:** Comparisons require reading multiple simulation summaries
- **Metric Calculation:** Computed on-demand for each comparison
- **Storage:** Results are cached in the database for fast retrieval
- **Pagination:** Use limits when listing many comparisons

## Troubleshooting

### No Results in Rankings

**Problem:** Some metrics show empty rankings
**Solution:** Check that simulations have completed and have valid summary data

### Inconsistent Rankings

**Problem:** Same simulation ranks differently across metrics
**Solution:** Expected behavior - different metrics measure different aspects of performance

### High Standard Deviation

**Problem:** Large variability in metric values
**Solution:** Indicates parameter sensitivity - consider narrower parameter ranges in future sweeps</content>
<parameter name="filePath">g:\code\Go\copilot\luckyfive\docs\comparison_guide.md