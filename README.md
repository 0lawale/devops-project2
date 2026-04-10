# DevOps Project 2 — URL Shortener with Docker Compose & GitHub Actions

A production-grade URL shortener built in Go, backed by Postgres and Redis,
fully containerised with Docker Compose and deployed automatically to AWS EC2
via a GitHub Actions CI/CD pipeline.

---

## Architecture

```
POST /shorten
      │
      ▼
  Go API ──▶ Postgres (permanent storage)
         ──▶ Redis (24hr cache)

GET /{code}
      │
      ▼
  Go API ──▶ Redis (cache hit → fast)
         ──▶ Postgres (cache miss → fallback)
              │
              ▼
         Redirect to original URL
```

---

## Tech Stack

| Tool | Purpose |
|------|---------|
| Go 1.23 | Application language |
| Postgres 16 | Persistent URL storage |
| Redis 7 | Cache layer (24hr TTL) |
| Docker Compose | Multi-service orchestration |
| GitHub Actions | CI/CD pipeline |
| AWS EC2 | Deployment target |

---

## CI/CD Pipeline

| Job | What it does |
|-----|-------------|
| **lint** | Runs golangci-lint for code quality |
| **test** | Runs unit tests with live Postgres and Redis services |
| **build** | Builds and pushes Docker image to Docker Hub tagged with commit SHA |
| **deploy** | SSHs into EC2 and runs latest container |

Every push to `main` triggers the full pipeline automatically.

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/shorten` | Shorten a URL |
| GET | `/{code}` | Redirect to original URL |

---

## Running Locally

**Prerequisites:** Docker, Docker Compose

```bash
git clone https://github.com/0lawale/devops-project2.git
cd devops-project2

docker compose up --build
```

Test it:

```bash
# Health check
curl http://localhost:8081/health

# Shorten a URL
curl -X POST http://localhost:8081/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://google.com"}'

# Redirect using the code returned above
curl -L http://localhost:8081/<code>
```

---

## Project Highlights

- **Cache-aside pattern** — Redis checked first on every read, falling back to Postgres on miss and repopulating cache automatically
- **Health-checked dependencies** — Docker Compose waits for Postgres and Redis to be healthy before starting the API
- **Persistent volumes** — data survives container restarts
- **Isolated networking** — all services communicate on a dedicated Docker bridge network
- **Commit-tagged images** — every Docker image tagged with Git commit SHA for full traceability