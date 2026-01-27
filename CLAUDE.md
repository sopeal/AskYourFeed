# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AskYourFeed is a full-stack application providing AI-powered Q&A over a user's Twitter/X feed. Users authenticate with their X username (validated via twitterapi.io), and the system ingests their feed data to answer questions using LLMs via OpenRouter.

## Tech Stack

- **Frontend**: React 19 + TypeScript + Vite + Tailwind CSS + Shadcn/UI
- **Backend**: Go 1.25 + Gin framework
- **Database**: PostgreSQL 16 with Row-Level Security (RLS)
- **External APIs**: twitterapi.io (Twitter data), OpenRouter (LLM)

## Development Commands

### Database (requires Docker)
```bash
make db-init       # Start PostgreSQL container and apply migrations
make db-start      # Start existing container
make db-stop       # Stop container
make db-reset      # Clean and reinitialize (destroys data)
make db-connect    # Connect via psql
make db-url        # Print connection string
```

### Backend (from `backend/` directory)
```bash
go run cmd/askyourfeed/main.go    # Start server on :8080
go test ./...                      # Run all tests
go test -v ./internal/services/   # Run specific package tests
go test -race -coverprofile=coverage.out ./...  # With race detection and coverage
```

### Frontend (from `frontend/` directory)
```bash
npm run dev      # Start Vite dev server
npm run build    # Build for production
npm run lint     # Run ESLint
npm run test     # Run Vitest (watch mode)
```

### E2E Tests (from project root)
```bash
npx playwright test              # Run all E2E tests
npx playwright test --ui         # Interactive UI mode
npx playwright test --debug      # Debug mode
npx playwright test e2e/register.spec.ts  # Single test file
```

## Architecture

### Backend Structure (Clean Architecture)
```
backend/
├── cmd/askyourfeed/main.go    # Entry point, dependency wiring
├── internal/
│   ├── handlers/              # HTTP handlers (controllers)
│   ├── services/              # Business logic layer
│   ├── repositories/          # Data access layer (PostgreSQL)
│   ├── middleware/            # Auth, CORS middleware
│   ├── dto/                   # Request/Response DTOs
│   ├── db/                    # Database connection utilities
│   └── testutil/              # Shared test utilities (helpers, mocks)
├── tests/
│   └── integration/           # Integration tests
└── pkg/logger/                # Shared logging package
```

**Testing**:
- Unit tests are co-located with code (e.g., `internal/services/*_test.go`)
- Integration tests are in `tests/integration/`
- Shared test utilities are in `internal/testutil/`

**API Routes** (all under `/api/v1`):
- `/auth/*` - Registration, login, logout
- `/session/current` - Current user session
- `/qa` - Q&A CRUD operations
- `/ingest` - Feed ingestion control
- `/following` - List followed authors

### Frontend Structure
```
frontend/src/
├── views/           # Page components (Dashboard, History, Login, Register)
├── components/      # Shared UI components (uses Shadcn/UI patterns)
├── hooks/           # Custom React hooks
├── schemas/         # Zod validation schemas
├── types/           # TypeScript type definitions
└── App.tsx          # Router configuration
```

**Routes**:
- `/login`, `/register` - Public auth pages
- `/` - Dashboard (protected)
- `/history` - Q&A history (protected)

## Code Conventions

### Go Backend
- Follow Clean Architecture: handlers → services → repositories
- Use interface-driven development with dependency injection
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Use table-driven tests with testify assertions
- Apply OpenTelemetry for tracing across service boundaries

### React Frontend
- Use React Hook Form + Zod for form validation
- Use TanStack Query for server state management
- Follow Shadcn/UI component patterns in `components/ui/`

## CI/CD

GitHub Actions runs on PRs to main/develop:
- Frontend: ESLint + Vitest with coverage
- Backend: golangci-lint + go test with coverage
