# Phase 0: Urgent & Ad-hoc Tasks

**Duration:** Ongoing (As needed)  
**Estimated Effort:** Variable  
**Team:** Dev Team  
**Status:** Active

---

## Overview

Phase 0 contains urgent, ad-hoc tasks that need immediate attention. These are high-priority items that don't fit into the regular phase structure but are critical for system stability, performance, or feature completion.

**Success Criteria:**
- ✅ All urgent issues resolved
- ✅ System stability improved
- ✅ Technical debt reduced
- ✅ User experience enhanced

---

## Task Breakdown

### Task 0.1: Migrate Sweep Configurations to Database
**Effort:** 12 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Migrate sweep configuration storage from JSON files in `docs/examples/sweeps/` to database records in `configs.db`. This provides better management, version control, and integration with the API.

**Acceptance Criteria:**
- [ ] Sweep configurations stored in database
- [ ] CRUD operations via API
- [ ] Existing examples migrated
- [ ] File-based storage removed
- [ ] Documentation updated

**Subtasks:**

1. **Create Database Schema for Sweeps** ✅
   - Add `sweeps` table to `configs.db` schema
   - Fields: id, name, description, config_json, created_at, updated_at, created_by
   - Create migration script in `migrations/`
   - Run migration and verify schema

2. **Generate sqlc Queries for Sweeps** ✅
   - Create `internal/store/queries/sweeps.sql` with CRUD queries
   - Generate types with `make generate`
   - Add Querier interface to store layer (generated as `sweeps.Querier`)
   - Test queries with sample data

3. **Implement SweepConfig Service** ✅
   - Create `internal/services/sweep_config.go`
   - Methods: CreateSweepConfig, GetSweepConfig, ListSweepConfigs, UpdateSweepConfig, DeleteSweepConfig
   - Integrate with existing config service patterns
   - Add validation using existing `pkg/sweep` package

4. **Create SweepConfig API Handlers**
   - Add `internal/handlers/sweep_configs.go`
   - Endpoints: POST /api/v1/sweep-configs, GET /api/v1/sweep-configs/:id, GET /api/v1/sweep-configs, PUT /api/v1/sweep-configs/:id, DELETE /api/v1/sweep-configs/:id
   - Wire handlers into router
   - Add proper error handling and validation

5. **Migrate Existing Examples**
   - Load all JSON files from `docs/examples/sweeps/`
   - Parse and validate each configuration
   - Insert into database via new service
   - Verify data integrity after migration
   - Create migration script for production deployment

6. **Update Sweep Generation to Use Database**
   - Modify sweep creation process to load config from database instead of files
   - Update `pkg/sweep/generator.go` if needed
   - Ensure backward compatibility during transition
   - Test with existing sweep workflows

7. **Remove File Dependencies**
   - Update documentation to reference database storage
   - Remove file loading code from services
   - Delete or archive `docs/examples/sweeps/` directory
   - Update any hardcoded file paths

8. **Add Comprehensive Tests**
   - Unit tests for SweepConfigService
   - Integration tests for API endpoints
   - Migration tests to ensure data preservation
   - Test coverage > 85% for new code

**Testing:**
- Test database CRUD operations
- Test API endpoints with various scenarios
- Test migration of existing configs
- Test sweep generation with database configs
- Verify no breaking changes to existing functionality

---

### Task 0.2: Database Performance Optimization
**Effort:** 8 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Optimize database queries and add indexes for better performance as data volume grows.

**Acceptance Criteria:**
- [ ] Query performance improved by 50%
- [ ] Proper indexes added
- [ ] Connection pooling configured
- [ ] Monitoring added

**Subtasks:**
1. Analyze slow queries
2. Add missing indexes
3. Implement query optimization
4. Add performance monitoring

---

### Task 0.3: Error Handling Standardization
**Effort:** 6 hours  
**Priority:** Medium  
**Assignee:** Dev 1

**Description:**
Standardize error handling across all services and handlers.

**Acceptance Criteria:**
- [ ] Consistent error responses
- [ ] Proper error logging
- [ ] User-friendly error messages
- [ ] Error codes documented

**Subtasks:**
1. Define error types
2. Update all handlers
3. Update error responses
4. Document error codes

---

## Phase 0 Checklist

- [ ] Task 0.1: Sweep database migration completed
- [ ] Task 0.2: Database optimization done
- [ ] Task 0.3: Error handling standardized
- [ ] All tests passing
- [ ] Documentation updated

---

## Metrics & KPIs

### Code Metrics
- **Test Coverage:** Maintain > 80%
- **Performance:** Query time < 100ms for common operations

---

## Deliverables Summary

1. **Database-backed Sweep Configs:** CRUD operations for sweep configurations
2. **Performance Improvements:** Optimized database queries
3. **Standardized Errors:** Consistent error handling across API

---

**Questions or Issues:**
Contact the development team or create an issue in the project tracker.