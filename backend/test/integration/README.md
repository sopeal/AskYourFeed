# Integration Tests

This directory contains integration tests for the AskYourFeed backend API.

## Prerequisites

1. **Docker**: Required for automatic database lifecycle management
2. **Go 1.25+**: For running the tests

## Automatic Database Management

The tests now include **automatic database lifecycle management**. When you run the tests:

1. A PostgreSQL Docker container is automatically started
2. The test database is created and initialized
3. All tests run against this isolated database
4. The container is automatically stopped and removed after tests complete

**No manual database setup is required!**

### Configuration

The automatic database management uses these defaults:

- **Container Name**: `askyourfeed_test_postgres`
- **Database Name**: `askyourfeed_test`
- **Port**: `5433` (to avoid conflicts with existing PostgreSQL instances)
- **User**: `postgres`
- **Password**: `postgres`

### Skipping Automatic Management

If you want to use an existing database instead:

```bash
export SKIP_DB_LIFECYCLE=true
export TEST_DATABASE_URL="postgres://user:password@localhost:5432/your_test_db?sslmode=disable"
```

## Running Tests

### Run All Integration Tests

Simply run:

```bash
cd backend
go test ./test/integration/... -v
```

The database will be automatically set up and torn down.

### Run Specific Test

```bash
cd backend
go test ./test/integration/... -v -run TestIngestStatusIntegration/HappyPath
```

### Run with Coverage

```bash
cd backend
go test ./test/integration/... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Quick Test (No Verbose Output)

```bash
cd backend
go test ./test/integration/...
```

## Test Structure

### Test Cases

The integration tests cover the following scenarios:

#### 1. Happy Path
- Tests successful retrieval of ingest status with completed runs
- Verifies `last_sync_at`, `recent_runs`, and proper ordering

#### 2. No Data
- Tests behavior when user has no ingest runs
- Verifies empty response structure

#### 3. Current Running
- Tests scenario with a currently running ingest
- Verifies `current_run` is populated correctly
- Ensures running ingest is not included in `recent_runs`

#### 4. Limit Parameter
- Tests default limit (10)
- Tests custom limit values
- Tests maximum limit (50)

#### 5. Edge Cases
- Limit exceeds maximum (51) → 400 Bad Request
- Limit is zero → 400 Bad Request
- Limit is negative → 400 Bad Request
- Limit is not a number → 400 Bad Request

#### 6. Error Cases
- Missing authorization header → 401 Unauthorized
- Invalid user ID format → 400 Bad Request

#### 7. Multiple Users
- Tests data isolation between different users
- Verifies each user only sees their own data

## Test Database Schema

The tests automatically create the required schema:

- `authors` table (global, no RLS)
- `ingest_runs` table (user-scoped, RLS enabled)
- Indexes and policies

The schema is created in the `applyMigrations` function and matches the production schema.

## Cleanup

Tests automatically clean up data between test runs using `TRUNCATE TABLE ingest_runs CASCADE`.

The Docker container is automatically removed after all tests complete.

## Troubleshooting

### Docker Not Running

If you get "Cannot connect to the Docker daemon" errors:

1. Ensure Docker is running: `sudo systemctl status docker`
2. Start Docker if needed: `sudo systemctl start docker`
3. Verify you have permissions: `docker ps`

### Port Already in Use

If port 5433 is already in use:

1. Check what's using the port: `lsof -i :5433`
2. Stop the conflicting service or change the test port in the code
3. Or use an existing database with `SKIP_DB_LIFECYCLE=true`

### Container Already Exists

If you see "container already exists" errors:

The tests automatically clean up old containers, but if needed:

```bash
docker stop askyourfeed_test_postgres
docker rm askyourfeed_test_postgres
```

### Tests Hang During Setup

If tests hang while waiting for PostgreSQL:

1. Check Docker logs: `docker logs askyourfeed_test_postgres`
2. Ensure you have the postgres:16-alpine image: `docker pull postgres:16-alpine`
3. Try manually starting the container to debug

### Permission Denied

If you get Docker permission errors:

```bash
sudo usermod -aG docker $USER
# Log out and back in for changes to take effect
```

## CI/CD Integration

The automatic database management works seamlessly in CI/CD pipelines. Simply run:

```bash
go test ./test/integration/... -v
```

The tests will handle all database setup and teardown automatically.

### Example GitHub Actions

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run Integration Tests
        run: |
          cd backend
          go test ./test/integration/... -v
```

### Example GitLab CI

```yaml
test:
  image: golang:1.25
  services:
    - docker:dind
  script:
    - cd backend
    - go test ./test/integration/... -v
```

## How It Works

The `TestMain` function in `ingest_status_test.go` handles the database lifecycle:

1. **Setup Phase**:
   - Checks if `SKIP_DB_LIFECYCLE` is set
   - Starts a PostgreSQL Docker container
   - Waits for PostgreSQL to be ready (up to 30 seconds)
   - Creates the test database

2. **Test Phase**:
   - Runs all tests
   - Each test gets a fresh database connection
   - Data is cleaned between test cases

3. **Teardown Phase**:
   - Stops the PostgreSQL container
   - Removes the container
   - Exits with the test result code

This ensures complete isolation and reproducibility for every test run.
