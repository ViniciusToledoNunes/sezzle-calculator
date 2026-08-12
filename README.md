# Sezzle Calculator

A full-stack calculator with a responsive React/TypeScript interface and a small Go REST API. It supports addition, subtraction, multiplication, division, exponentiation, square root, and percentage conversion.

The project favors explicit validation, isolated business logic, accessible interaction, and a deployment path that remains easy to review.

## Highlights

- Seven arithmetic operations, including all optional assignment operations
- Validation in both the browser and API, with structured error responses
- Division-by-zero, invalid-domain, overflow, and malformed-request handling
- Responsive, keyboard-friendly UI with loading and screen-reader feedback
- Session-only calculation history (nothing is persisted)
- Unit and integration-style handler/component tests
- Single-container production build and automated CI

## Architecture

```text
frontend/                    React + TypeScript + Vite
  src/api/                   HTTP client and API error translation
  src/domain/                Operation definitions and formatting
  src/App.tsx                Calculator workflow and presentation

backend/                     Go, standard library only
  internal/calculator/       Pure arithmetic domain logic
  internal/httpapi/          JSON transport, validation, CORS, SPA serving
  cmd/server/                Configuration and graceful HTTP server
```

The frontend calls one operation-oriented REST endpoint. In development, Vite proxies `/api` to Go; in the production image, Go serves the compiled frontend and API from the same origin.

## Prerequisites

- Go 1.23+
- Node.js 22+ and npm 10+
- Docker (optional)

## Run locally

Start the backend:

```bash
cd backend
go run ./cmd/server
```

In another terminal, start the frontend:

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173). The API listens on `http://localhost:8080`.

Backend configuration:

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP listening port |
| `ALLOWED_ORIGIN` | `http://localhost:5173` | Allowed browser origin in development |
| `STATIC_DIR` | unset | Optional compiled frontend directory |

To point the frontend at a separately hosted API, set `VITE_API_URL` when building or running Vite.

## Run with Docker

Build and run the entire application as one container:

```bash
docker build -t sezzle-calculator .
docker run --rm -p 8080:8080 sezzle-calculator
```

Open [http://localhost:8080](http://localhost:8080). The image uses a non-root runtime user and includes a health check.

## API

### Health check

```http
GET /healthz
```

```json
{"status":"ok"}
```

### Calculate

```http
POST /api/v1/calculate
Content-Type: application/json
```

Binary operation example:

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"multiply","operands":[6,7]}'
```

```json
{"operation":"multiply","operands":[6,7],"result":42}
```

Unary operation example:

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"square_root","operands":[81]}'
```

Supported operation values:

| Operation | Operands | Meaning |
| --- | ---: | --- |
| `add` | 2 | `a + b` |
| `subtract` | 2 | `a - b` |
| `multiply` | 2 | `a × b` |
| `divide` | 2 | `a ÷ b` |
| `power` | 2 | `aᵇ` |
| `square_root` | 1 | `√a` |
| `percentage` | 1 | `a ÷ 100` |

Errors have a stable machine-readable code and a user-facing message:

```json
{
  "error": {
    "code": "division_by_zero",
    "message": "division by zero is not allowed."
  }
}
```

Malformed requests return `400 Bad Request`; valid requests with impossible operands return `422 Unprocessable Entity`.

## Tests and coverage

Backend:

```bash
cd backend
go test ./... -coverprofile coverage.out
go tool cover -func coverage.out
```

Frontend:

```bash
cd frontend
npm ci
npm run typecheck
npm run test:coverage
npm run build
```

Current measured coverage:

- Backend: **94.1% statements** overall; calculator domain **100%**; HTTP layer **100%**
- Frontend: **96.78% statements**, **84.81% branches**, **93.75% functions**
- Frontend suite: **8 tests**; backend uses table-driven tests across domain and HTTP edge cases

See the detailed [coverage report](docs/coverage.md). CI repeats type checking, tests, coverage generation, and the production frontend build on every push and pull request.

## Design decisions and assumptions

- **One calculation endpoint:** the operation is data, so a single endpoint keeps validation and response contracts consistent without seven nearly identical handlers.
- **Pure domain package:** arithmetic has no HTTP concerns, making edge cases fast to test and allowing another transport to reuse it.
- **`float64` arithmetic:** appropriate for a general calculator. Exact financial decimal math is intentionally out of scope and would use a decimal representation in a money-moving system.
- **Percentage semantics:** `percentage(25)` returns `0.25`; the UI describes this as “convert to decimal.”
- **Defense in depth:** the UI catches common mistakes immediately, while the API remains authoritative and independently validates arity, operation names, domains, finite inputs, and finite results.
- **No database:** recent results are intentionally held only in React state for the current tab.
- **Standard-library backend:** fewer dependencies, a small attack surface, and simple review/build behavior.
- **Accessible UI:** visible labels, 44px+ controls, clear focus states, semantic status/error announcements, adequate contrast, and reduced-motion support are built in.

## AI usage

AI-assisted implementation was permitted by the assignment. The prompt record is included in [PROMPTS.md](PROMPTS.md).
