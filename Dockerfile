FROM golang:1.25.3-alpine AS build

ENV CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -trimpath -ldflags "-s -w" -o /app/main cmd/api/main.go

FROM alpine:3.22 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
RUN adduser -D -H -u 10001 appuser && chown -R appuser:appuser /app
USER 10001
EXPOSE 8080
CMD ["./main"]
