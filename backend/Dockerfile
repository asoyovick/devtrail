FROM golang:1.22.2-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /devtrail-api ./cmd/api

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -H -u 10001 devtrail
COPY --from=build /devtrail-api /usr/local/bin/devtrail-api
COPY migrations ./migrations
COPY configs/config.example.env ./configs/config.example.env
USER devtrail
EXPOSE 8080
CMD ["devtrail-api"]
