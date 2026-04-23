## 1. The Service Architecture (The Template)

Each Go microservice should follow a **Hexagonal (Ports and Adapters)** or **Clean Architecture** structure. This ensures your business logic doesn't get tangled with your database or API logic.

### Internal Folder Structure

```
/my-distributed-app
├── go.work                # The workspace definition
├── docker-compose.yaml    # Orchestrates the local environment
├── common/                # Shared internal library
│   ├── go.mod
│   ├── auth/              # JWT validation logic
│   └── logger/            # slog configurations
├── services/
│   ├── gateway/           # Entry point (Router)
│   │   ├── go.mod
│   │   └── main.go
│   ├── identity/          # Auth & User management
│   │   ├── go.mod
│   │   └── main.go
│   └── backoffice/        # Admin logic
│       ├── go.mod
│       └── main.go
└── api/                   # Shared API definitions (Protobuf/OpenAPI)
```

---

## 2. Communication & Integration

- **API Gateway:** Use a gateway (like **Nginx** or a simple Go service using **Go-Cloud-Proxy**) to route `/api/v1/users` to the Identity service and `/api/v1/admin` to the Backoffice logic.
- **Contract First:** Define your APIs using **Protocol Buffers (.proto files)**. Even if you use REST, Protobuf provides a strict schema that serves as documentation and allows you to generate Go code for both clients and servers.
- **Shared Identity:** A centralized **Auth Service** that issues **JWTs (JSON Web Tokens)**. Every other service should be able to validate these tokens using a shared public key without needing to call the Auth service every time.

---

## 3. The Infrastructure Stack

To keep the template "cloud-native," you should include these in your environment:

| Component             | Recommended Tool         | Purpose                                                                        |
| :-------------------- | :----------------------- | :----------------------------------------------------------------------------- |
| **Orchestration**     | **Docker Compose**       | To spin up all services, DBs, and brokers with one command.                    |
| **Database**          | **PostgreSQL**           | Use separate schemas or databases for different microservices.                 |
| **Migration**         | **golang-migrate**       | Never manually edit your DB; use versioned migration files.                    |
| **Service Discovery** | **CoreDNS / Docker DNS** | Allows services to find each other by name (e.g., `http://auth-service:8080`). |

---

## 4. Mandatory Shared Features

Your template isn't "production-ready" until these cross-cutting concerns are handled:

- **Structured Logging:** Use the standard `slog` package. Logs should be JSON formatted so they can be parsed by tools like ELK or Grafana Loki.
- **Environment Config:** Use **Viper** or **Godotenv** to manage secrets and URLs across different environments (Dev vs. Prod).
- **Graceful Shutdown:** Ensure your Go services listen for `SIGTERM` signals to close database connections and finish ongoing requests before exiting.

---

## 5. First Development Milestone

1.  Set up a **Monorepo** (one git repository) with a `services/` folder.
2.  Create an `auth-service` that handles `Login` and `Signup`.
3.  Create a `gateway` that forwards requests.
4.  Write a `docker-compose.yml` that launches both, along with a Postgres instance.

Once you can log in through the gateway and receive a JWT, your "Distributed Template" is officially alive.

---

## 6. Current Local Runtime

The repository currently runs:

- `auth-service` on `:8081`
- `backoffice` on `:8082`
- `gateway` on `:8080`
- Postgres on `:5432`

### Gateway routes

- `GET /healthz`
- `POST /api/v1/auth/signup`
- `POST /api/v1/auth/login`
- `GET /api/v1/admin/status` (proxied to `backoffice`)

### Backoffice environment variables

- `SERVICE_NAME` (default: `backoffice`)
- `LOG_LEVEL` (default: `info`)
- `HTTP_ADDR` (default: `:8082`)
- `ADMIN_ROUTE_PREFIX` (default: `/api/v1/admin`)

### Run locally

```bash
docker compose up --build
```
