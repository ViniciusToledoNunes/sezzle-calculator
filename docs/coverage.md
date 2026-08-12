# Coverage report

Coverage was generated locally from the repository source after the implementation.

## Backend (Go)

Command:

```bash
cd backend
go test ./... -coverprofile coverage.out
go tool cover -func coverage.out
```

| Package | Statement coverage |
| --- | ---: |
| `internal/calculator` | 100.0% |
| `internal/httpapi` | 100.0% |
| `cmd/server` | 72.0% |
| **Overall** | **94.1%** |

The executable package includes the intentionally uncalled `main` process boundary. Its testable server lifecycle covers configuration, successful graceful shutdown, and listener failures; the `run` function itself is 93.3% covered.

## Frontend (React/TypeScript)

Command:

```bash
cd frontend
npm run test:coverage
```

| Metric | Coverage |
| --- | ---: |
| Statements | 96.78% |
| Branches | 84.81% |
| Functions | 93.75% |
| Lines | 96.78% |

The generated HTML reports are intentionally excluded from Git. Run the commands above to recreate `backend/coverage.out` and `frontend/coverage/index.html`.
