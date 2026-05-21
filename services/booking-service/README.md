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
