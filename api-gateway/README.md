# API Gateway

The API Gateway exposes HTTP routes and calls the internal gRPC services.

## Booking Routes

- `POST /api/bookings`
- `GET /api/bookings/{bookingID}`
- `GET /api/users/{userID}/bookings`
- `POST /api/bookings/{bookingID}/confirm`
- `POST /api/bookings/{bookingID}/cancel`

## Runtime Configuration

- `GATEWAY_HTTP_ADDRESS`: HTTP listen address, defaults to `:8080`
- `BOOKING_GRPC_TARGET`: Booking Service gRPC target, defaults to `localhost:50053`
- `USER_GRPC_TARGET`: User Service gRPC target, defaults to `localhost:50051`
- `INVENTORY_GRPC_TARGET`: Inventory Service gRPC target, defaults to `localhost:50052`

## Demo routes

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/cars`
- `GET /api/cars`
- `GET /api/cars/{carID}`
- `POST /api/bookings`
- `POST /api/bookings/{bookingID}/confirm`
- `POST /api/bookings/{bookingID}/cancel`
- `GET /api/users/{userID}/bookings`

Run the gateway from the repository root and open `http://localhost:8080` for
the browser demo console.
