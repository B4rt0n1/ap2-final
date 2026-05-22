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
