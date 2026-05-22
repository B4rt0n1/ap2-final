# Demo Runbook

Run commands from the repository root unless a command changes directory.

## Start infrastructure

```cmd
docker compose up -d booking-postgres user-postgres inventory-postgres nats redis
docker compose --profile tools run --rm booking-migrate
docker compose --profile tools run --rm user-migrate
docker compose --profile tools run --rm inventory-migrate
```

## Start services

Terminal 1:

```cmd
set USER_DATABASE_URL=postgres://user:user@localhost:55434/users?sslmode=disable
set USER_GRPC_ADDRESS=:50051
set NATS_URL=nats://localhost:4222
go run ./services/user-service/cmd/server
```

Terminal 2:

```cmd
cd services\inventory-service
set INVENTORY_DATABASE_URL=postgres://inventory:inventory@localhost:55435/inventory?sslmode=disable
set INVENTORY_GRPC_ADDRESS=:50052
set REDIS_ADDRESS=localhost:6379
set NATS_URL=nats://localhost:4222
go run ./cmd/server
```

Terminal 3:

```cmd
set BOOKING_DATABASE_URL=postgres://booking:booking@localhost:55433/booking?sslmode=disable
set BOOKING_GRPC_ADDRESS=:50053
set USER_GRPC_ADDRESS=localhost:50051
set INVENTORY_GRPC_ADDRESS=localhost:50052
set NATS_URL=nats://localhost:4222
go run ./services/booking-service/cmd/server
```

Terminal 4:

```cmd
set GATEWAY_HTTP_ADDRESS=:8080
set BOOKING_GRPC_TARGET=localhost:50053
set USER_GRPC_TARGET=localhost:50051
set INVENTORY_GRPC_TARGET=localhost:50052
go run ./api-gateway/cmd/server
```

Open the browser console at `http://localhost:8080`.

## Postman flow

Set `baseUrl` to `http://localhost:8080`.

1. Register a user.

   `POST {{baseUrl}}/api/v1/auth/register`

   ```json
   {
     "first_name": "Aron",
     "last_name": "Demo",
     "email": "aron.demo@example.com",
     "password": "secret123",
     "phone": "+77000000000"
   }
   ```

   Copy `user.id` from the response into a Postman variable named `userId`.

2. Add a car.

   `POST {{baseUrl}}/api/cars`

   ```json
   {
     "brand": "Toyota",
     "model": "Camry",
     "year": 2025,
     "category": "sedan",
     "price_per_day": 85
   }
   ```

   Copy `car.id` from the response into `carId`.

3. Read the fleet.

   `GET {{baseUrl}}/api/cars`

   Run `GET {{baseUrl}}/api/cars/{{carId}}` twice when explaining Redis:
   the first lookup loads from Postgres and the next lookup can come from cache.

4. Create a booking.

   `POST {{baseUrl}}/api/bookings`

   ```json
   {
     "user_id": "{{userId}}",
     "car_id": "{{carId}}",
     "start_date": "2026-06-01",
     "end_date": "2026-06-04"
   }
   ```

   Copy `booking.id` from the response into `bookingId`.

5. Confirm it.

   `POST {{baseUrl}}/api/bookings/{{bookingId}}/confirm`

6. List renter bookings.

   `GET {{baseUrl}}/api/users/{{userId}}/bookings`

7. Cancel an unconfirmed demo booking when needed.

   `POST {{baseUrl}}/api/bookings/{{bookingId}}/cancel`

## Email

To send a real `booking.created` email from User Service, set `SMTP_HOST`,
`SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, and `SMTP_FROM` before starting
User Service.
