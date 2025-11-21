# Copilot Instructions for LuckyFive Project

## Project Overview
LuckyFive is a lottery analysis system built in Go that processes lottery results, performs statistical analysis, and provides insights for lottery games. The system includes data import, simulation capabilities, and REST API endpoints.

## Architecture
- **Language**: Go 1.21+
- **Database**: SQLite with sqlc for type-safe queries
- **HTTP Framework**: Chi router
- **Logging**: slog (structured logging)
- **Configuration**: Environment variables with config package

## Code Structure
```
├── cmd/           # Executables (api, admin, migrate, worker)
├── internal/      # Private application code
│   ├── config/    # Configuration management
│   ├── handlers/  # HTTP handlers
│   ├── middleware/# HTTP middleware
│   ├── models/    # Data models
│   ├── services/  # Business logic
│   └── store/     # Database layer
├── migrations/    # Database migrations
├── pkg/           # Public packages
└── data/          # Data files and temp storage
```

## Development Guidelines

### Be meticulous
- Be meticulous in every change: run the test suite, update relevant documentation, and validate migrations or DB schema changes before committing. Small, well-tested incremental changes are preferred over large, risky edits. When modifying code, double-check error handling, edge cases, and resource cleanup (files, DB connections, goroutines).

**Documentation discipline**
- After completing a task or making a change, update all relevant documentation (design docs, implementation plans, change logs, and task lists). Don't assume someone else will update docs — the author of the change is responsible.
- Before marking a task or change as done, read the entire related document(s) end-to-end (e.g. `design_doc_v2.md`, `implementation_plan.md`, `plans/phase1_tasks.md`) to ensure no related item was left pending and the documentation accurately reflects the new behavior.
- If the change affects public APIs, data formats, migrations, or operational procedures, add explicit migration notes and a short 'what changed' entry in the repository README or a changelog file.


### Code Style
- **Write only in English**: All code, comments, and documentation must be in English
- **Use `any` instead of `interface{}`**: Prefer `any` for type parameters and empty interfaces
- **Use range loops**: Instead of `for i := 0; i < limit; i++`, use `for i := range limit`
- Follow Go conventions and idioms
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions small and focused on single responsibility

### Error Handling
- Return errors instead of panicking
- Use structured logging with slog
- Provide meaningful error messages
- Handle edge cases gracefully

### Testing
- **Never use testify**: Use only stdlib testing and `errors.Is` for error checking
- **Include failure tests**: When creating tests, include tests for failure scenarios as well
- Write unit tests for all business logic
- Create tests for all packages, services, and handlers created
- Use table-driven tests for multiple test cases
- Mock external dependencies when necessary
- Aim for good test coverage

### Database
- Use sqlc for type-safe database queries
- Follow migration patterns for schema changes
- Use transactions for multi-step operations
- Validate data before database operations

### API Design
- RESTful endpoints with consistent naming
- JSON request/response format
- Proper HTTP status codes
- Input validation and sanitization

## Common Patterns

### Service Layer
```go
type MyService struct {
    db     *store.DB
    logger *slog.Logger
}

func NewMyService(db *store.DB, logger *slog.Logger) *MyService {
    return &MyService{db: db, logger: logger}
}
```

### HTTP Handlers
```go
func MyHandler(service *services.MyService, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Handle request
    }
}
```
- **Separation of Concerns**: Handlers should only handle HTTP request/response (parsing, validation, formatting)
- **Business Logic**: Services should process the actual business logic and data operations
- Parse request data, validate input, call service methods, format responses

### Error Responses
```go
WriteError(w, http.StatusBadRequest, APIError{
    Code:    "error_code",
    Message: "Human readable message",
})
```

## File Upload Handling
- Store uploaded files temporarily in `data/temp/`
- Generate unique artifact IDs for tracking
- Clean up files after successful processing
- Validate file types and sizes

## Import Process
1. Upload XLSX file via `/api/v1/results/upload`
2. Receive artifact_id in response
3. Call `/api/v1/results/import` with artifact_id
4. System processes file and imports data
5. Returns import statistics

## Testing Strategy
- **Testing Framework**: Use only Go standard library testing (`testing` package)
- **Error Checking**: Use `errors.Is` for error comparison, never `assert` from testify
- **Coverage**: Create tests for all packages, services, and handlers
- **Failure Scenarios**: Always include tests for failure cases, not just success paths
- Unit tests for services and handlers
- Integration tests for API endpoints
- Use in-memory databases for testing
- Mock external services when needed
- Table-driven tests for multiple scenarios

## Performance Considerations
- Use database indexes for frequently queried columns
- Implement pagination for large result sets
- Cache expensive computations when appropriate
- Monitor memory usage for large file processing

## Security
- Validate all input data
- Use prepared statements (handled by sqlc)
- Implement proper authentication/authorization if needed
- Sanitize file uploads

## Deployment
- Use environment variables for configuration
- Include health check endpoints
- Log structured information for monitoring
- Handle graceful shutdowns

## Contributing
- **Task Documentation**: When finishing a task, update the task documentation
- **Reference Documents**: Use `design_doc_v2.md` and `implementation_plan.md` as references
- Follow the established patterns and conventions
- Write tests for new functionality
- Update documentation as needed
- Ensure code builds and tests pass before committing

## Documentation
- **Design Reference**: `design_doc_v2.md` contains system design and architecture decisions
- **Implementation Reference**: `implementation_plan.md` contains task breakdown and implementation details
- Update task status in implementation documents when completing work
- Keep documentation synchronized with code changes