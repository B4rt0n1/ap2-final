# Car Rental Project Overview

## Goal

The project is a car rental system built as microservices for the AP2 final.
Users register, browse cars, create rental bookings, and receive booking email
notifications.

The final system should demonstrate:

- Clean Architecture
- at least three microservices and one API Gateway
- gRPC endpoints and service-to-service gRPC calls
- PostgreSQL databases with migrations and transactions
- Redis cache usage
- NATS message queue usage
- SMTP email sending
- unit and integration tests
- GitHub Actions CI

## Repositories

The project uses two GitHub repositories.

### `B4rt0n1/final_proto`

This repository owns gRPC contracts:

- `.proto` files
- Buf configuration
- generated Go protobuf and gRPC packages

### `B4rt0n1/ap2-final`

This repository owns the running system:

- microservices
- API Gateway
- database migrations
- Docker Compose infrastructure
- Redis and NATS integration
- tests
- GitHub Actions

## Services

The system has three microservices and one gateway.

### User Service

The User Service owns user data and authentication flows:

- register and login
- user profile
- password and email verification flows
- driver license data
- SMTP email integration or email worker integration

### Inventory Service

The Inventory Service owns car data:

- add, update, delete, and get cars
- list and search cars
- car availability
- rental pricing
- Redis cache for car reads

Good Redis use cases:

- cache `GetCarById`
- cache `ListCars`
- invalidate cache after car update or delete

### Booking Service

The Booking Service owns rental reservations:

- create and update bookings
- confirm or cancel bookings
- start and end rentals
- calculate booking cost
- apply discounts
- booking status
- booking issue reports

The Booking Service is the main service-to-service gRPC flow:

- it calls User Service to check the user
- it calls Inventory Service to check car availability
- it calls Inventory Service to get car pricing

### API Gateway

The API Gateway is the client entrypoint.

Examples:

- `/auth` and `/users` forward to User Service
- `/cars` forwards to Inventory Service
- `/bookings` forwards to Booking Service

Every team member should implement their own gateway area.

## Runtime Architecture

```text
Client
  |
  v
API Gateway
  |
  | gRPC
  +--> User Service ------> PostgreSQL
  |
  +--> Inventory Service -> PostgreSQL
  |                      -> Redis
  |
  +--> Booking Service --> PostgreSQL
                         -> User Service via gRPC
                         -> Inventory Service via gRPC
                         -> NATS
                              |
                              v
                         Email Worker
                              |
                              v
                            SMTP
```

## Main Booking Flow

1. A user registers through the API Gateway.
2. User Service stores the user in PostgreSQL.
3. The client requests cars through the API Gateway.
4. Inventory Service serves car data from Redis when possible and falls back to PostgreSQL.
5. The client creates a booking through the API Gateway.
6. Booking Service checks the user through gRPC.
7. Booking Service checks car availability and pricing through Inventory Service gRPC.
8. Booking Service saves the booking in PostgreSQL inside a transaction.
9. Booking Service publishes a booking event to NATS.
10. An email worker consumes the event and sends an SMTP email.

## Message Queue

Use NATS for asynchronous events and jobs.

Booking events can include:

- `booking.created`
- `booking.confirmed`
- `booking.cancelled`

Email sending should react asynchronously to queue messages instead of blocking
the main booking request.

## Data and Transactions

Each service owns its own data boundary.

Suggested main tables:

- User Service: `users`
- Inventory Service: `cars`
- Booking Service: `bookings`

Booking Service transactions should be visible in flows that change booking
state, such as create, confirm, cancel, start, end, or discount updates.

## Team Split

### Participant 1

- User Service
- proto repository support
- email flow and SMTP worker
- API Gateway user and auth routes

### Participant 2

- Booking Service
- booking transactions
- service-to-service gRPC calls to User and Inventory
- NATS booking events
- API Gateway booking routes
- shared CI or infrastructure work when needed

### Participant 3

- Inventory Service
- cars database and migrations
- basic Redis cache
- API Gateway car routes

## Final Demo

A strong demo should show:

1. user registration
2. car creation and car listing
3. Redis miss and Redis hit for car reads
4. booking creation
5. Booking Service gRPC calls to User and Inventory
6. booking database transaction
7. NATS booking event
8. email sent after the event
9. passing tests and GitHub Actions
