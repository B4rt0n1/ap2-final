# AP2 Final Car Rental

This repository contains the car rental microservices, API gateway, and runtime infrastructure for the final project.

The Protobuf contracts and generated Go gRPC packages live in the separate `B4rt0n1/final_proto` repository. This project imports the generated packages from:

- `github.com/B4rt0n1/final_proto/gen/go/user/v1`
- `github.com/B4rt0n1/final_proto/gen/go/inventory/v1`
- `github.com/B4rt0n1/final_proto/gen/go/booking/v1`

See [docs/project-overview.md](docs/project-overview.md) for the project architecture, team split, and demo flow.

## Local Infrastructure

Start shared local infrastructure:

```bash
docker compose up -d booking-postgres user-postgres inventory-postgres nats redis
```

Apply service migrations:

```bash
docker compose --profile tools run --rm booking-migrate
docker compose --profile tools run --rm user-migrate
docker compose --profile tools run --rm inventory-migrate
```

The default local databases are exposed on `localhost:55433` for Booking,
`localhost:55434` for User, and `localhost:55435` for Inventory. Copy values
from `.env.example` into your shell environment or local `.env` before starting
the gRPC processes. Configure the `SMTP_*` values to send booking emails from
the User Service worker.

## API Gateway

The API Gateway exposes HTTP routes on `:8080` by default and calls Booking
Service gRPC at `localhost:50053` and User Service gRPC at `localhost:50051`.

```bash
go run ./api-gateway/cmd/server
```

## CI

GitHub Actions runs Go tests, builds the Booking Service, Inventory Service,
and API Gateway entrypoints, and validates the Docker Compose configuration on
pushes and pull requests.
