-- Migration: 006_migrate_sweep_examples.sql
-- Migrates existing sweep configuration examples from docs/examples/sweeps/ to database

-- Up migration

-- Insert alpha_range.json
INSERT OR IGNORE INTO sweeps (name, description, config_json, created_by) VALUES (
    'alpha_range_sweep',
    'Sweep alpha parameter from 0.0 to 1.0 in steps of 0.1',
    '{"name":"alpha_range_sweep","description":"Sweep alpha parameter from 0.0 to 1.0 in steps of 0.1","base_recipe":{"version":"1.0","name":"advanced","parameters":{"alpha":0.1,"sim_prev_max":100,"sim_count":1000,"scorer_type":"frequency"}},"parameters":[{"name":"alpha","type":"range","values":{"min":0.0,"max":1.0,"step":0.1}}]}',
    'migration'
);

-- Insert exponential.json
INSERT OR IGNORE INTO sweeps (name, description, config_json, created_by) VALUES (
    'exponential_sweep',
    'Exponential sweep for parameters that vary by orders of magnitude',
    '{"name":"exponential_sweep","description":"Exponential sweep for parameters that vary by orders of magnitude","base_recipe":{"version":"1.0","name":"advanced","parameters":{"alpha":0.001,"sim_prev_max":100,"sim_count":1000,"scorer_type":"frequency"}},"parameters":[{"name":"alpha","type":"exponential","values":{"base":10,"start":-3,"end":0}},{"name":"sim_prev_max","type":"exponential","values":{"base":10,"start":1,"end":3}}]}',
    'migration'
);

-- Insert multi_param.json
INSERT OR IGNORE INTO sweeps (name, description, config_json, created_by) VALUES (
    'multi_param_sweep',
    'Multi-parameter sweep with alpha, sim_prev_max, and constraints',
    '{"name":"multi_param_sweep","description":"Multi-parameter sweep with alpha, sim_prev_max, and constraints","base_recipe":{"version":"1.0","name":"advanced","parameters":{"alpha":0.1,"sim_prev_max":100,"sim_count":1000,"scorer_type":"frequency"}},"parameters":[{"name":"alpha","type":"range","values":{"min":0.0,"max":0.8,"step":0.1}},{"name":"sim_prev_max","type":"discrete","values":{"values":[50,100,200,500,1000]}},{"name":"sim_count","type":"exponential","values":{"base":10,"start":2,"end":4}}],"constraints":[{"type":"max","parameters":["alpha"],"value":0.8}]}',
    'migration'
);

-- Down migration
-- DELETE FROM sweeps WHERE created_by = 'migration';