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
docker compose up -d booking-postgres nats redis
```

Apply Booking Service migrations:

```bash
docker compose --profile tools run --rm booking-migrate
```

The default booking database for local development is exposed on `localhost:5433`.
Copy values from `.env.example` into your shell environment or local `.env`
before starting the booking gRPC process.
