# Booking Service

The booking service owns rental reservations and their lifecycle.

## Layout

- `cmd/server`: gRPC service entrypoint
- `internal/domain`: booking entities, statuses, and domain errors
- `internal/usecase`: booking application flows
- `internal/repository`: persistence ports
- `internal/transport/grpc`: generated-contract adapters
- `internal/infrastructure`: database and external service adapters
- `migrations`: booking database schema migrations

## Runtime Configuration

The booking gRPC process reads:

- `BOOKING_DATABASE_URL`: required PostgreSQL DSN
- `BOOKING_GRPC_ADDRESS`: server listen address, defaults to `:50053`
- `USER_GRPC_ADDRESS`: User Service gRPC address, defaults to `localhost:50051`
- `INVENTORY_GRPC_ADDRESS`: Inventory Service gRPC address, defaults to `localhost:50052`
- `NATS_URL`: NATS URL, defaults to `nats://localhost:4222`

## Local Infrastructure

From the repository root:

```bash
docker compose up -d booking-postgres nats redis
docker compose --profile tools run --rm booking-migrate
```

Then start the booking gRPC process with `BOOKING_DATABASE_URL` set to the DSN
from `.env.example`.
